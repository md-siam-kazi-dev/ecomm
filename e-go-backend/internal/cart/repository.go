package cart

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// CartRepository is the persistence boundary for carts.
type CartRepository interface {
	FindUserCart(ctx context.Context, userID string) (Cart, bool, error)
	FindSessionCart(ctx context.Context, sessionID string) (Cart, bool, error)
	CreateCart(ctx context.Context, userID, sessionID, currency string) (Cart, error)
	MergeCarts(ctx context.Context, fromCartID, intoCartID string) error
	MarkCartStatus(ctx context.Context, cartID, status string) error

	ResolveVariant(ctx context.Context, productID, variantID string) (VariantResolution, error)
	ListItems(ctx context.Context, cartID string) ([]ItemRecord, error)
	UpsertItem(ctx context.Context, cartID, variantID string, quantity int, unitPrice float64, maxQuantity int) error
	FindItem(ctx context.Context, cartID, itemID string) (ItemRecord, error)
	SetItemQuantity(ctx context.Context, cartID, itemID string, quantity int) error
	RemoveItem(ctx context.Context, cartID, itemID string) error
	ClearItems(ctx context.Context, cartID string) error
}

// Repository persists carts in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

const cartColumns = `id, COALESCE(user_id::text, ''), COALESCE(session_id, ''), currency`

func scanCart(row pgx.Row) (Cart, error) {
	var c Cart
	err := row.Scan(&c.ID, &c.UserID, &c.SessionID, &c.Currency)
	return c, err
}

// FindUserCart returns the caller's active cart, if any.
func (r *Repository) FindUserCart(ctx context.Context, userID string) (Cart, bool, error) {
	c, err := scanCart(r.db.QueryRow(ctx, `
		SELECT `+cartColumns+`
		FROM carts
		WHERE user_id = $1 AND status = 'active'
		ORDER BY updated_at DESC
		LIMIT 1`, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, false, nil
	}
	if err != nil {
		return Cart{}, false, apperr.ErrInternal.WithCause(err)
	}
	return c, true, nil
}

// FindSessionCart returns the guest cart for a session id, if any.
func (r *Repository) FindSessionCart(ctx context.Context, sessionID string) (Cart, bool, error) {
	c, err := scanCart(r.db.QueryRow(ctx, `
		SELECT `+cartColumns+`
		FROM carts
		WHERE session_id = $1 AND status = 'active'
		ORDER BY updated_at DESC
		LIMIT 1`, sessionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, false, nil
	}
	if err != nil {
		return Cart{}, false, apperr.ErrInternal.WithCause(err)
	}
	return c, true, nil
}

// CreateCart inserts a new active cart bound to a user and/or a session id.
func (r *Repository) CreateCart(ctx context.Context, userID, sessionID, currency string) (Cart, error) {
	c, err := scanCart(r.db.QueryRow(ctx, `
		INSERT INTO carts (user_id, session_id, status, currency)
		VALUES ($1, $2, 'active', $3)
		RETURNING `+cartColumns, nullableUUID(userID), nullableString(sessionID), currency))
	if err != nil {
		return Cart{}, apperr.ErrInternal.WithCause(err)
	}
	return c, nil
}

// MergeCarts moves every line from one cart into another (summing quantities,
// capped), empties the source cart and marks it converted — atomically.
func (r *Repository) MergeCarts(ctx context.Context, fromCartID, intoCartID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		INSERT INTO cart_items (cart_id, variant_id, quantity, unit_price)
		SELECT $2, variant_id, quantity, unit_price
		FROM cart_items
		WHERE cart_id = $1
		ON CONFLICT (cart_id, variant_id)
		DO UPDATE SET
			quantity = LEAST($3, cart_items.quantity + EXCLUDED.quantity),
			unit_price = EXCLUDED.unit_price,
			updated_at = NOW()`,
		fromCartID, intoCartID, MaxLineQuantity); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, fromCartID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE carts SET status = 'converted', updated_at = NOW() WHERE id = $1`, fromCartID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	if _, err := tx.Exec(ctx, `UPDATE carts SET updated_at = NOW() WHERE id = $1`, intoCartID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// MarkCartStatus transitions a cart's status (e.g. to 'converted' at checkout).
func (r *Repository) MarkCartStatus(ctx context.Context, cartID, status string) error {
	if _, err := r.db.Exec(ctx, `UPDATE carts SET status = $2, updated_at = NOW() WHERE id = $1`, cartID, status); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// ResolveVariant resolves a product (+ optional variant) to a concrete variant,
// its price and available stock. Only live (active + approved) products resolve.
func (r *Repository) ResolveVariant(ctx context.Context, productID, variantID string) (VariantResolution, error) {
	var (
		res      VariantResolution
		status   string
		approval string
	)
	err := r.db.QueryRow(ctx, `
		SELECT v.id, p.id, p.name, v.price::float8,
		       COALESCE((SELECT SUM(i.quantity) FROM inventory i WHERE i.variant_id = v.id), 0)::bigint,
		       p.status::text, p.approval_status::text
		FROM product_variants v
		JOIN products p ON p.id = v.product_id
		WHERE p.id = $1 AND ($2 = '' OR v.id::text = $2)
		ORDER BY v.created_at
		LIMIT 1`, productID, variantID,
	).Scan(&res.VariantID, &res.ProductID, &res.Name, &res.UnitPrice, &res.Available, &status, &approval)
	if errors.Is(err, pgx.ErrNoRows) {
		return VariantResolution{}, errProductUnavailable
	}
	if err != nil {
		return VariantResolution{}, apperr.ErrInternal.WithCause(err)
	}
	if status != "active" || approval != "approved" {
		return VariantResolution{}, errProductUnavailable
	}
	return res, nil
}

const itemColumns = `
	ci.id, p.id, v.id, p.name, p.slug,
	COALESCE(img.url, ''), COALESCE(img.alt_text, ''),
	COALESCE(attrs.label, ''),
	ci.unit_price::float8, ci.quantity,
	COALESCE((SELECT SUM(i.quantity) FROM inventory i WHERE i.variant_id = v.id), 0)::bigint`

const itemJoins = `
	FROM cart_items ci
	JOIN product_variants v ON v.id = ci.variant_id
	JOIN products p ON p.id = v.product_id
	LEFT JOIN LATERAL (
		SELECT url, alt_text FROM product_images pi
		WHERE pi.product_id = p.id
		ORDER BY pi.is_primary DESC, pi.sort_order
		LIMIT 1
	) img ON TRUE
	LEFT JOIN LATERAL (
		SELECT string_agg(pa.display_name || ': ' || COALESCE(av.display_value, av.value), ', ' ORDER BY pa.name) AS label
		FROM variant_attribute_values vav
		JOIN attribute_values av ON av.id = vav.attribute_value_id
		JOIN product_attributes pa ON pa.id = av.attribute_id
		WHERE vav.variant_id = v.id
	) attrs ON TRUE`

func scanItem(row pgx.Row) (ItemRecord, error) {
	var it ItemRecord
	err := row.Scan(&it.ID, &it.ProductID, &it.VariantID, &it.Name, &it.Slug,
		&it.ImageURL, &it.ImageAlt, &it.VariantLabel, &it.UnitPrice, &it.Quantity, &it.Available)
	return it, err
}

// ListItems returns a cart's lines with display + stock data (oldest first).
func (r *Repository) ListItems(ctx context.Context, cartID string) ([]ItemRecord, error) {
	rows, err := r.db.Query(ctx, `SELECT `+itemColumns+itemJoins+`
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at, ci.id`, cartID)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	items := make([]ItemRecord, 0)
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return items, nil
}

// FindItem returns a single cart line (scoped to the cart).
func (r *Repository) FindItem(ctx context.Context, cartID, itemID string) (ItemRecord, error) {
	it, err := scanItem(r.db.QueryRow(ctx, `SELECT `+itemColumns+itemJoins+`
		WHERE ci.cart_id = $1 AND ci.id = $2`, cartID, itemID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ItemRecord{}, errItemNotFound
	}
	if err != nil {
		return ItemRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return it, nil
}

// UpsertItem adds a line or increments an existing one, capping the line at
// maxQuantity and refreshing the stored unit price.
func (r *Repository) UpsertItem(ctx context.Context, cartID, variantID string, quantity int, unitPrice float64, maxQuantity int) error {
	if quantity > maxQuantity {
		quantity = maxQuantity
	}
	if _, err := r.db.Exec(ctx, `
		INSERT INTO cart_items (cart_id, variant_id, quantity, unit_price)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (cart_id, variant_id)
		DO UPDATE SET
			quantity = LEAST($5, cart_items.quantity + EXCLUDED.quantity),
			unit_price = EXCLUDED.unit_price,
			updated_at = NOW()`,
		cartID, variantID, quantity, unitPrice, maxQuantity); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// SetItemQuantity sets a line's quantity (0 removes the line).
func (r *Repository) SetItemQuantity(ctx context.Context, cartID, itemID string, quantity int) error {
	if quantity <= 0 {
		return r.RemoveItem(ctx, cartID, itemID)
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE cart_items SET quantity = $3, updated_at = NOW()
		WHERE cart_id = $1 AND id = $2`, cartID, itemID, quantity)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return errItemNotFound
	}
	return nil
}

// RemoveItem deletes a single line from a cart.
func (r *Repository) RemoveItem(ctx context.Context, cartID, itemID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1 AND id = $2`, cartID, itemID)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return errItemNotFound
	}
	return nil
}

// ClearItems removes every line from a cart.
func (r *Repository) ClearItems(ctx context.Context, cartID string) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// ---- helpers --------------------------------------------------------------

func nullableUUID(id string) any {
	if id == "" {
		return nil
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil
	}
	return parsed
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
