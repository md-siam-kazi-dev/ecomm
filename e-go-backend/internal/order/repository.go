package order

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

const pgUniqueViolation = "23505"

// errOrderNumberTaken is an internal signal so the service can retry.
var errOrderNumberTaken = errors.New("order number taken")

// OrderRepository is the persistence boundary for orders.
type OrderRepository interface {
	ResolveVariants(ctx context.Context, items []OrderItemInput) ([]VariantResolution, error)
	CartItemsForUser(ctx context.Context, userID string) ([]ItemRequest, error)
	ConvertUserCart(ctx context.Context, userID string) error
	CreateOrder(ctx context.Context, p OrderParams) (Order, error)
	DeleteOrder(ctx context.Context, orderID string) error
	OrdersForUser(ctx context.Context, userID string, page, limit int) ([]Order, int64, error)
	OrderByID(ctx context.Context, userID, id string) (Order, error)

	CreatePayment(ctx context.Context, orderID, provider string, amount float64, currency string) (Payment, error)
	PaymentByReference(ctx context.Context, reference string) (Payment, error)
	MarkPaid(ctx context.Context, reference, providerPaymentID string) (Payment, bool, error)
	InsertPaymentTransaction(ctx context.Context, paymentID, providerPaymentID string, amount float64, currency string) error
	ConfirmPaidOrder(ctx context.Context, reference string) error
}

// Repository persists orders in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// ResolveVariants loads the server-side price, vendor and stock for each line.
func (r *Repository) ResolveVariants(ctx context.Context, items []OrderItemInput) ([]VariantResolution, error) {
	out := make([]VariantResolution, 0, len(items))
	for _, item := range items {
		var (
			res      VariantResolution
			status   string
			approval string
		)
		err := r.db.QueryRow(ctx, `
			SELECT v.id, p.id, p.vendor_id, p.name, v.sku, v.price::float8,
			       COALESCE((SELECT SUM(i.quantity) FROM inventory i WHERE i.variant_id = v.id), 0)::bigint,
			       p.status::text, p.approval_status::text
			FROM product_variants v
			JOIN products p ON p.id = v.product_id
			WHERE p.id = $1 AND ($2 = '' OR v.id::text = $2)
			ORDER BY v.created_at
			LIMIT 1`, item.ProductID, item.VariantID,
		).Scan(&res.VariantID, &res.ProductID, &res.VendorID, &res.ProductName, &res.ProductSKU,
			&res.UnitPrice, &res.Available, &status, &approval)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errUnavailable
		}
		if err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		if status != "active" || approval != "approved" {
			return nil, errUnavailable
		}
		out = append(out, res)
	}
	return out, nil
}

// CartItemsForUser loads the lines of the user's active server-side cart so a
// checkout can be created from the cart instead of a client-supplied payload.
func (r *Repository) CartItemsForUser(ctx context.Context, userID string) ([]ItemRequest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id::text, ci.variant_id::text, ci.quantity
		FROM carts c
		JOIN cart_items ci ON ci.cart_id = c.id
		JOIN product_variants v ON v.id = ci.variant_id
		JOIN products p ON p.id = v.product_id
		WHERE c.user_id = $1 AND c.status = 'active'
		ORDER BY ci.created_at, ci.id`, userID)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	items := make([]ItemRequest, 0)
	for rows.Next() {
		var item ItemRequest
		if err := rows.Scan(&item.ProductID, &item.VariantID, &item.Quantity); err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return items, nil
}

// ConvertUserCart marks the user's active cart 'converted' after a successful
// checkout, so it is no longer reused (the cart is emptied server-side).
func (r *Repository) ConvertUserCart(ctx context.Context, userID string) error {
	if _, err := r.db.Exec(ctx, `
		UPDATE carts SET status = 'converted', updated_at = NOW()
		WHERE user_id = $1 AND status = 'active'`, userID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// CreateOrder inserts the order, items, address, history and decrements stock
// in a single transaction. Concurrent stock races are caught by the guarded
// inventory update.
func (r *Repository) CreateOrder(ctx context.Context, p OrderParams) (Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Order{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		orderID   string
		createdAt time.Time
	)
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (order_number, user_id, status, currency, subtotal, discount_total, tax_total, shipping_total, grand_total, notes)
		VALUES ($1, $2, 'pending', $3, $4, 0, $5, $6, $7, '')
		RETURNING id, created_at`,
		p.OrderNumber, p.UserID, p.Currency, p.Subtotal, p.TaxTotal, p.ShippingTotal, p.GrandTotal,
	).Scan(&orderID, &createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return Order{}, errOrderNumberTaken
		}
		return Order{}, apperr.ErrInternal.WithCause(err)
	}

	items := make([]OrderItem, 0, len(p.Items))
	for i, item := range p.Items {
		quantity := p.Quantities[i]
		lineTotal := round2(item.UnitPrice * float64(quantity))

		if _, err := tx.Exec(ctx, `
			INSERT INTO order_items
				(order_id, vendor_id, product_id, variant_id, product_name, product_sku, quantity, unit_price, discount, tax, subtotal, total)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, 0, $9, $10)`,
			orderID, item.VendorID, item.ProductID, item.VariantID, item.ProductName, item.ProductSKU,
			quantity, item.UnitPrice, lineTotal, lineTotal,
		); err != nil {
			return Order{}, apperr.ErrInternal.WithCause(err)
		}

		var locationID string
		var stockAfter int
		err := tx.QueryRow(ctx, `
			UPDATE inventory SET quantity = quantity - $2, updated_at = NOW()
			WHERE id = (SELECT id FROM inventory WHERE variant_id = $1 ORDER BY quantity DESC LIMIT 1)
			  AND quantity >= $2
			RETURNING location_id, quantity`, item.VariantID, quantity).Scan(&locationID, &stockAfter)
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, errInsufficientStock
		}
		if err != nil {
			return Order{}, apperr.ErrInternal.WithCause(err)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO inventory_movements
				(variant_id, location_id, movement_type, quantity, quantity_before, quantity_after, reference_type, reference_id, notes, created_by)
			VALUES ($1, $2, 'sale', $3, $4, $5, 'order', $6, 'Order sale', $7)`,
			item.VariantID, locationID, quantity, stockAfter+quantity, stockAfter, orderID, p.UserID,
		); err != nil {
			return Order{}, apperr.ErrInternal.WithCause(err)
		}

		items = append(items, OrderItem{
			ProductID:   item.ProductID,
			VariantID:   item.VariantID,
			ProductName: item.ProductName,
			ProductSKU:  item.ProductSKU,
			Quantity:    quantity,
			UnitPrice:   item.UnitPrice,
			LineTotal:   lineTotal,
		})
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO order_addresses
			(order_id, address_type, full_name, phone, address_line_1, address_line_2, city, state, postal_code, country)
		VALUES ($1, 'shipping', $2, $3, $4, $5, $6, $7, $8, $9)`,
		orderID, p.Shipping.FullName, p.Shipping.Phone, p.Shipping.Line1, nullable(p.Shipping.Line2),
		p.Shipping.City, p.Shipping.State, p.Shipping.PostalCode, p.Shipping.Country,
	); err != nil {
		return Order{}, apperr.ErrInternal.WithCause(err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO order_status_history (order_id, old_status, new_status, changed_by, notes)
		VALUES ($1, NULL, 'pending', $2, 'Order placed')`, orderID, p.UserID); err != nil {
		return Order{}, apperr.ErrInternal.WithCause(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Order{}, apperr.ErrInternal.WithCause(err)
	}

	return Order{
		ID:            orderID,
		OrderNumber:   p.OrderNumber,
		Status:        StatusPending,
		Currency:      p.Currency,
		Subtotal:      p.Subtotal,
		ShippingTotal: p.ShippingTotal,
		TaxTotal:      p.TaxTotal,
		GrandTotal:    p.GrandTotal,
		CreatedAt:     createdAt,
		Items:         items,
		Shipping:      p.Shipping,
	}, nil
}

// DeleteOrder compensates an order whose payment could not be started: it
// restores the stock taken and removes the order (and any payment rows) in one
// transaction, so a failed payment never leaves an orphan order.
func (r *Repository) DeleteOrder(ctx context.Context, orderID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `SELECT COALESCE(variant_id::text, ''), quantity FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	type line struct {
		variantID string
		quantity  int
	}
	var lines []line
	for rows.Next() {
		var l line
		if err := rows.Scan(&l.variantID, &l.quantity); err != nil {
			rows.Close()
			return apperr.ErrInternal.WithCause(err)
		}
		lines = append(lines, l)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	for _, l := range lines {
		if l.variantID == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
			UPDATE inventory SET quantity = quantity + $2, updated_at = NOW()
			WHERE id = (SELECT id FROM inventory WHERE variant_id = $1 ORDER BY quantity DESC LIMIT 1)`,
			l.variantID, l.quantity,
		); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
	}

	for _, stmt := range []string{
		`DELETE FROM payment_transactions WHERE payment_id IN (SELECT id FROM payments WHERE order_id = $1)`,
		`DELETE FROM payments WHERE order_id = $1`,
		`DELETE FROM order_status_history WHERE order_id = $1`,
		`DELETE FROM order_addresses WHERE order_id = $1`,
		`DELETE FROM order_items WHERE order_id = $1`,
	} {
		if _, err := tx.Exec(ctx, stmt, orderID); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM orders WHERE id = $1`, orderID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	return tx.Commit(ctx)
}

// OrdersForUser returns one page of the user's orders (newest first).
func (r *Repository) OrdersForUser(ctx context.Context, userID string, page, limit int) ([]Order, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM orders WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []Order{}, 0, nil
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, order_number, status::text, currency, subtotal::float8, shipping_total::float8,
		       tax_total::float8, grand_total::float8, created_at
		FROM orders WHERE user_id = $1
		ORDER BY created_at DESC, id
		LIMIT $2 OFFSET $3`, userID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	orders, err := scanOrders(rows)
	if err != nil {
		return nil, 0, err
	}
	ids := make([]string, 0, len(orders))
	for _, o := range orders {
		ids = append(ids, o.ID)
	}
	if err := r.attachItems(ctx, orders, ids); err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// OrderByID returns one order owned by the user, with items and address.
func (r *Repository) OrderByID(ctx context.Context, userID, id string) (Order, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, order_number, status::text, currency, subtotal::float8, shipping_total::float8,
		       tax_total::float8, grand_total::float8, created_at
		FROM orders WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return Order{}, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	orders, err := scanOrders(rows)
	if err != nil {
		return Order{}, err
	}
	if len(orders) == 0 {
		return Order{}, errNotFound
	}
	if err := r.attachItems(ctx, orders, []string{orders[0].ID}); err != nil {
		return Order{}, err
	}

	var addr AddressParams
	err = r.db.QueryRow(ctx, `
		SELECT full_name, COALESCE(phone, ''), address_line_1, COALESCE(address_line_2, ''),
		       city, COALESCE(state, ''), COALESCE(postal_code, ''), country
		FROM order_addresses WHERE order_id = $1 AND address_type = 'shipping'
		ORDER BY created_at LIMIT 1`, orders[0].ID,
	).Scan(&addr.FullName, &addr.Phone, &addr.Line1, &addr.Line2, &addr.City, &addr.State, &addr.PostalCode, &addr.Country)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Order{}, apperr.ErrInternal.WithCause(err)
	}
	orders[0].Shipping = addr
	return orders[0], nil
}

// CreatePayment inserts a pending payment row for an order.
func (r *Repository) CreatePayment(ctx context.Context, orderID, provider string, amount float64, currency string) (Payment, error) {
	var p Payment
	err := r.db.QueryRow(ctx, `
		INSERT INTO payments (order_id, provider, status, amount, currency)
		VALUES ($1, $2, 'pending', $3, $4)
		RETURNING id, order_id, provider, COALESCE(provider_payment_id, ''), status::text, amount::float8, currency, verified_at`,
		orderID, provider, amount, currency,
	).Scan(&p.ID, &p.OrderID, &p.Provider, &p.ProviderPaymentID, &p.Status, &p.Amount, &p.Currency, &p.VerifiedAt)
	if err != nil {
		return Payment{}, apperr.ErrInternal.WithCause(err)
	}
	return p, nil
}

const paymentColumns = `
	p.id, p.order_id, o.order_number, p.provider, COALESCE(p.provider_payment_id, ''),
	p.status::text, p.amount::float8, p.currency, p.verified_at`

func scanPayment(s scanner) (Payment, error) {
	var p Payment
	err := s.Scan(&p.ID, &p.OrderID, &p.Reference, &p.Provider, &p.ProviderPaymentID,
		&p.Status, &p.Amount, &p.Currency, &p.VerifiedAt)
	return p, err
}

// PaymentByReference loads the payment for the order whose number matches.
func (r *Repository) PaymentByReference(ctx context.Context, reference string) (Payment, error) {
	p, err := scanPayment(r.db.QueryRow(ctx, `
		SELECT `+paymentColumns+`
		FROM payments p JOIN orders o ON o.id = p.order_id
		WHERE o.order_number = $1
		ORDER BY p.created_at DESC LIMIT 1`, reference))
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, errNotFound
	}
	if err != nil {
		return Payment{}, apperr.ErrInternal.WithCause(err)
	}
	return p, nil
}

// MarkPaid transitions a pending/processing payment to paid (idempotent).
func (r *Repository) MarkPaid(ctx context.Context, reference, providerPaymentID string) (Payment, bool, error) {
	p, err := scanPayment(r.db.QueryRow(ctx, `
		UPDATE payments p SET status = 'paid',
		       provider_payment_id = COALESCE(NULLIF($2, ''), p.provider_payment_id),
		       verified_at = NOW(), updated_at = NOW()
		FROM orders o
		WHERE o.id = p.order_id AND o.order_number = $1 AND p.status IN ('pending', 'processing')
		RETURNING `+paymentColumns, reference, providerPaymentID))
	if err == nil {
		return p, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, false, apperr.ErrInternal.WithCause(err)
	}
	// No row updated: already paid (or unknown reference).
	existing, getErr := r.PaymentByReference(ctx, reference)
	if getErr != nil {
		return Payment{}, false, getErr
	}
	return existing, false, nil
}

// InsertPaymentTransaction records the captured payment.
func (r *Repository) InsertPaymentTransaction(ctx context.Context, paymentID, providerPaymentID string, amount float64, currency string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO payment_transactions (payment_id, transaction_type, amount, provider_transaction_id, status)
		VALUES ($1, 'payment', $2, NULLIF($3, ''), 'paid')`, paymentID, amount, providerPaymentID)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// ConfirmPaidOrder moves a pending order to confirmed exactly once.
func (r *Repository) ConfirmPaidOrder(ctx context.Context, reference string) error {
	var orderID string
	err := r.db.QueryRow(ctx, `
		UPDATE orders SET status = 'confirmed', updated_at = NOW()
		WHERE order_number = $1 AND status = 'pending'
		RETURNING id`, reference).Scan(&orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // already confirmed (idempotent)
	}
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if _, err := r.db.Exec(ctx, `
		INSERT INTO order_status_history (order_id, old_status, new_status, notes)
		VALUES ($1, 'pending', 'confirmed', 'Payment confirmed')`, orderID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// ---- helpers --------------------------------------------------------------

type scanner interface{ Scan(dest ...any) error }

func scanOrders(rows pgx.Rows) ([]Order, error) {
	orders := make([]Order, 0)
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.Status, &o.Currency, &o.Subtotal,
			&o.ShippingTotal, &o.TaxTotal, &o.GrandTotal, &o.CreatedAt); err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return orders, nil
}

func (r *Repository) attachItems(ctx context.Context, orders []Order, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT oi.order_id, oi.id, COALESCE(oi.product_id::text, ''), COALESCE(oi.variant_id::text, ''),
		       oi.product_name, COALESCE(oi.product_sku, ''), oi.quantity, oi.unit_price::float8, oi.total::float8,
		       COALESCE(img.url, ''),
		       rv.id, COALESCE(rv.rating, 0), COALESCE(rv.title, ''), COALESCE(rv.content, ''), rv.created_at
		FROM order_items oi
		LEFT JOIN LATERAL (
			SELECT url FROM product_images pi
			WHERE pi.product_id = oi.product_id
			ORDER BY pi.is_primary DESC, pi.sort_order
			LIMIT 1
		) img ON TRUE
		LEFT JOIN reviews rv ON rv.order_item_id = oi.id
		WHERE oi.order_id::text = ANY(string_to_array($1, ','))
		ORDER BY oi.created_at`, strings.Join(ids, ","))
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	byOrder := make(map[string][]OrderItem, len(orders))
	reviewByID := make(map[string]*OrderItemReview)
	for rows.Next() {
		var (
			orderID   string
			item      OrderItem
			reviewID  *string
			rating    int
			title     string
			comment   string
			createdAt *time.Time
		)
		if err := rows.Scan(&orderID, &item.OrderItemID, &item.ProductID, &item.VariantID, &item.ProductName,
			&item.ProductSKU, &item.Quantity, &item.UnitPrice, &item.LineTotal, &item.ImageURL,
			&reviewID, &rating, &title, &comment, &createdAt); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
		if reviewID != nil {
			item.Reviewed = true
			review := &OrderItemReview{ID: *reviewID, Rating: rating, Title: title, Comment: comment}
			if createdAt != nil {
				review.CreatedAt = *createdAt
			}
			item.Review = review
			reviewByID[*reviewID] = review
		}
		byOrder[orderID] = append(byOrder[orderID], item)
	}
	if err := rows.Err(); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	if err := r.attachReviewImages(ctx, reviewByID); err != nil {
		return err
	}
	for i := range orders {
		if items, ok := byOrder[orders[i].ID]; ok {
			orders[i].Items = items
		} else {
			orders[i].Items = []OrderItem{}
		}
	}
	return nil
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// attachReviewImages groups each review's photo URLs onto its OrderItemReview.
func (r *Repository) attachReviewImages(ctx context.Context, byID map[string]*OrderItemReview) error {
	if len(byID) == 0 {
		return nil
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	rows, err := r.db.Query(ctx, `
		SELECT review_id, image_url FROM review_images
		WHERE review_id::text = ANY(string_to_array($1, ','))
		ORDER BY review_id, sort_order`, strings.Join(ids, ","))
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			reviewID string
			url      string
		)
		if err := rows.Scan(&reviewID, &url); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
		if review, ok := byID[reviewID]; ok {
			review.Images = append(review.Images, url)
		}
	}
	if err := rows.Err(); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

func nullable(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}
