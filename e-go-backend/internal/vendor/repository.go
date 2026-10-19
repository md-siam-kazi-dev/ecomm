package vendor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/internal/notify"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// lowStockThreshold matches the storefront's "low stock" definition.
const lowStockThreshold = 10

// errNoActiveStore is returned when the authenticated user has no active store.
var errNoActiveStore = apperr.New(http.StatusForbidden, "NO_ACTIVE_STORE", "You do not have an active store")

// errStoreNotActive is returned when the resolved store is not active.
var errStoreNotActive = apperr.New(http.StatusForbidden, "STORE_NOT_ACTIVE", "This store is not active")

// VendorRepository is the persistence boundary for the vendor module.
type VendorRepository interface {
	ResolveStore(ctx context.Context, userID, storeID string) (StoreScope, error)
	Dashboard(ctx context.Context, vendorID string) (Summary, error)
	ListProducts(ctx context.Context, vendorID string, q ListQuery) ([]ProductRecord, int64, error)
	CreateProduct(ctx context.Context, vendorID string, p ProductParams) (ProductDetail, error)
	UpdateProduct(ctx context.Context, vendorID, productID string, p ProductParams) (ProductDetail, error)
	GetProduct(ctx context.Context, vendorID, productID string) (ProductDetail, error)
	ListOrders(ctx context.Context, vendorID string, q ListQuery) ([]OrderRecord, int64, error)
	OrderDetail(ctx context.Context, vendorID, orderID string) (OrderDetailRecord, error)
	UpsertShipmentEvent(ctx context.Context, vendorID, orderID, actorID string, p ShipmentParams) error
	SetOrderStatus(ctx context.Context, vendorID, orderID, actorID, status, note string) error
	Revenue(ctx context.Context, vendorID string) (RevenueSummary, error)
	ListInventory(ctx context.Context, vendorID string, q ListQuery) ([]InventoryRecord, int64, error)
	ListInventoryMovements(ctx context.Context, vendorID string, q ListQuery) ([]InventoryMovementRecord, int64, error)
	SetReorderLevel(ctx context.Context, vendorID, variantID, userID string, level int) (InventoryRecord, error)
	AdjustStock(ctx context.Context, vendorID, variantID, userID string, quantity int, note string) (InventoryRecord, error)
	ExportOrders(ctx context.Context, vendorID string, q ExportQuery) ([]ExportOrderRow, error)
	ExportInventory(ctx context.Context, vendorID string, q ListQuery) ([]ExportInventoryRow, error)
	ListReviews(ctx context.Context, vendorID string, q ListQuery) ([]VendorReviewRecord, int64, error)
	ReviewSummary(ctx context.Context, vendorID string) (ReviewSummaryRecord, error)
	ReplyToReview(ctx context.Context, vendorID, reviewID, reply string) (VendorReviewRecord, error)
	ListQuestions(ctx context.Context, vendorID string, q ListQuery) ([]VendorQuestionRecord, int64, error)
	AnswerQuestion(ctx context.Context, vendorID, questionID, userID, answer string) (VendorQuestionRecord, error)
}

// errProductNotFound is returned when a product does not exist or is not owned
// by the vendor.
var errProductNotFound = apperr.New(http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")

// errSKUTaken is returned when the supplied SKU is already used.
var errSKUTaken = apperr.New(http.StatusConflict, "SKU_TAKEN", "That SKU is already in use")

// Repository persists vendor data in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// ResolveStore maps the authenticated user to the store they are operating.
// This is the store-scoped authorization gate: the user must be an active
// member of an active store. When storeID is empty the caller's store is chosen.
func (r *Repository) ResolveStore(ctx context.Context, userID, storeID string) (StoreScope, error) {
	var sc StoreScope
	err := r.db.QueryRow(ctx, `
		SELECT s.id, s.vendor_id, s.name, s.status::text, m.role::text
		FROM store_memberships m
		JOIN stores s ON s.id = m.store_id
		WHERE m.user_id = $1 AND m.status = 'active'
		  AND ($2 = '' OR s.id::text = $2)
		ORDER BY (s.status = 'active') DESC, s.created_at DESC
		LIMIT 1`, userID, storeID,
	).Scan(&sc.StoreID, &sc.VendorID, &sc.Name, &sc.Status, &sc.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return StoreScope{}, errNoActiveStore
	}
	if err != nil {
		return StoreScope{}, apperr.ErrInternal.WithCause(err)
	}
	if sc.Status != "active" {
		return StoreScope{}, errStoreNotActive
	}
	return sc, nil
}

// Dashboard aggregates the vendor's KPIs. Every figure is scoped to the vendor
// and computed in PostgreSQL.
func (r *Repository) Dashboard(ctx context.Context, vendorID string) (Summary, error) {
	var s Summary
	err := r.db.QueryRow(ctx, `
		SELECT
			COALESCE((
				SELECT SUM(oi.total)::float8
				FROM order_items oi
				JOIN orders o ON o.id = oi.order_id
				WHERE oi.vendor_id = $1
				  AND o.status NOT IN ('cancelled', 'returned', 'refunded')
			), 0),
			(SELECT COUNT(DISTINCT oi.order_id) FROM order_items oi WHERE oi.vendor_id = $1),
			(SELECT COUNT(*) FROM products p WHERE p.vendor_id = $1),
			COALESCE((
				SELECT COUNT(*) FROM (
					SELECT v.id
					FROM product_variants v
					JOIN products p ON p.id = v.product_id
					LEFT JOIN inventory i ON i.variant_id = v.id
					WHERE p.vendor_id = $1
					GROUP BY v.id
					HAVING COALESCE(SUM(i.quantity - i.reserved_quantity), 0) <= $2
				) low
			), 0),
			(SELECT COUNT(DISTINCT oi.order_id)
			 FROM order_items oi
			 JOIN orders o ON o.id = oi.order_id
			 WHERE oi.vendor_id = $1
			   AND o.status IN ('pending', 'confirmed', 'processing', 'packed'))
	`, vendorID, lowStockThreshold).Scan(
		&s.Revenue, &s.OrderCount, &s.ProductCount, &s.LowStockCount, &s.PendingFulfilment,
	)
	if err != nil {
		return Summary{}, apperr.ErrInternal.WithCause(err)
	}
	return s, nil
}

// ListProducts returns one page of the vendor's products plus the total count.
func (r *Repository) ListProducts(ctx context.Context, vendorID string, q ListQuery) ([]ProductRecord, int64, error) {
	args := []any{vendorID}
	search := ""
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		place := "$" + strconv.Itoa(len(args))
		search = " AND (p.name ILIKE " + place + " OR p.slug ILIKE " + place + ")"
	}

	var total int64
	if err := r.db.QueryRow(ctx,
		`SELECT count(*) FROM products p WHERE p.vendor_id = $1`+search, args...,
	).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []ProductRecord{}, 0, nil
	}

	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)
	query := `
		SELECT p.id, p.name, p.slug, p.status::text, p.approval_status::text, p.created_at,
		       COALESCE(p.review_note, ''), p.submitted_at, p.reviewed_at,
		       COALESCE((SELECT v.price FROM product_variants v WHERE v.product_id = p.id ORDER BY v.created_at LIMIT 1), 0)::float8,
		       COALESCE((SELECT SUM(i.quantity) FROM inventory i JOIN product_variants v ON v.id = i.variant_id WHERE v.product_id = p.id), 0)::bigint,
		       COALESCE((SELECT img.url FROM product_images img WHERE img.product_id = p.id ORDER BY img.is_primary DESC, img.sort_order LIMIT 1), '')
		FROM products p
		WHERE p.vendor_id = $1` + search + `
		ORDER BY p.created_at DESC, p.id
		LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]ProductRecord, 0, q.Limit)
	for rows.Next() {
		var rec ProductRecord
		if err := rows.Scan(
			&rec.ID, &rec.Name, &rec.Slug, &rec.Status, &rec.ApprovalStatus, &rec.CreatedAt,
			&rec.ReviewNote, &rec.SubmittedAt, &rec.ReviewedAt, &rec.Price, &rec.Stock, &rec.ImageURL,
		); err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	return records, total, nil
}

// ListOrders returns one page of orders containing the vendor's items.
func (r *Repository) ListOrders(ctx context.Context, vendorID string, q ListQuery) ([]OrderRecord, int64, error) {
	args := []any{vendorID}
	statusFilter := ""
	if q.Status != "" {
		args = append(args, q.Status)
		statusFilter = " AND o.status = $" + strconv.Itoa(len(args)) + "::order_status"
	}

	var total int64
	if err := r.db.QueryRow(ctx, `
		SELECT count(DISTINCT oi.order_id)
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.vendor_id = $1`+statusFilter, args...,
	).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []OrderRecord{}, 0, nil
	}

	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)
	query := `
		SELECT o.id, o.order_number, o.status::text, o.created_at,
		       COALESCE(SUM(oi.quantity), 0)::bigint,
		       COALESCE(SUM(oi.total), 0)::float8
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.vendor_id = $1` + statusFilter + `
		GROUP BY o.id
		ORDER BY o.created_at DESC, o.id
		LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]OrderRecord, 0, q.Limit)
	for rows.Next() {
		var rec OrderRecord
		if err := rows.Scan(&rec.OrderID, &rec.OrderNumber, &rec.Status, &rec.CreatedAt, &rec.ItemCount, &rec.Total); err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	return records, total, nil
}

// normalizeSearch trims and bounds a search term.
func normalizeSearch(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > maxSearchLen {
		s = s[:maxSearchLen]
	}
	return s
}

// ---- inventory ------------------------------------------------------------

// errStockBelowReserved is returned when an adjustment would drop stock below
// the units already reserved for open orders.
var errStockBelowReserved = apperr.New(
	http.StatusUnprocessableEntity, "STOCK_BELOW_RESERVED",
	"Stock cannot be lower than the reserved quantity",
)

const inventoryFrom = `
	FROM product_variants v
	JOIN products p ON p.id = v.product_id
	LEFT JOIN inventory i ON i.variant_id = v.id
	LEFT JOIN product_categories pc ON pc.product_id = p.id
	LEFT JOIN categories c ON c.id = pc.category_id
	LEFT JOIN LATERAL (
		SELECT url FROM product_images pi
		WHERE pi.product_id = p.id
		ORDER BY pi.is_primary DESC, pi.sort_order
		LIMIT 1
	) img ON TRUE`

const inventoryColumns = `v.id, p.id, p.name, v.sku, COALESCE(c.name, ''), COALESCE(img.url, ''),
	COALESCE(SUM(i.quantity), 0)::bigint, COALESCE(SUM(i.reserved_quantity), 0)::bigint,
	COALESCE(MAX(i.reorder_level), 0)::bigint`

// buildInventoryFilter returns the shared WHERE/HAVING clauses and positional
// args for the inventory count and page queries. Status is in|low|out.
func buildInventoryFilter(vendorID string, q ListQuery) (where, having string, args []any) {
	args = []any{vendorID}
	where = " WHERE p.vendor_id = $1"

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		place := "$" + strconv.Itoa(len(args))
		where += " AND (p.name ILIKE " + place + " OR v.sku ILIKE " + place + ")"
	}

	switch q.Status {
	case "out":
		having = " HAVING COALESCE(SUM(i.quantity), 0) = 0"
	case "low":
		args = append(args, lowStockThreshold)
		having = " HAVING COALESCE(SUM(i.quantity), 0) > 0 AND COALESCE(SUM(i.quantity), 0) <= $" + strconv.Itoa(len(args))
	case "in":
		args = append(args, lowStockThreshold)
		having = " HAVING COALESCE(SUM(i.quantity), 0) > $" + strconv.Itoa(len(args))
	case "unpublished":
		where += " AND (p.status <> 'active' OR p.approval_status <> 'approved')"
	}

	return where, having, args
}

// ListInventory returns one page of the vendor's inventory (per variant) plus
// the total matching count.
func (r *Repository) ListInventory(ctx context.Context, vendorID string, q ListQuery) ([]InventoryRecord, int64, error) {
	where, having, args := buildInventoryFilter(vendorID, q)

	var total int64
	if err := r.db.QueryRow(ctx, `
		SELECT count(*) FROM (
			SELECT v.id`+inventoryFrom+where+`
			GROUP BY v.id`+having+`
		) t`, args...).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []InventoryRecord{}, 0, nil
	}

	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.db.Query(ctx, `
		SELECT `+inventoryColumns+inventoryFrom+where+`
		GROUP BY v.id, p.id, c.name, img.url`+having+`
		ORDER BY p.name, v.sku
		LIMIT $`+strconv.Itoa(len(args)+1)+` OFFSET $`+strconv.Itoa(len(args)+2), listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]InventoryRecord, 0, q.Limit)
	for rows.Next() {
		var rec InventoryRecord
		if err := rows.Scan(
			&rec.VariantID, &rec.ProductID, &rec.Name, &rec.SKU, &rec.Category,
			&rec.ImageURL, &rec.Quantity, &rec.Reserved, &rec.ReorderLevel,
		); err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	return records, total, nil
}

// AdjustStock sets a variant's on-hand stock and records the movement. The
// variant must belong to the vendor.
func (r *Repository) AdjustStock(ctx context.Context, vendorID, variantID, userID string, quantity int, note string) (InventoryRecord, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var ownerID, productName, sku, vendorUserID string
	err = tx.QueryRow(ctx, `
		SELECT p.vendor_id, p.name, COALESCE(v.sku, ''), vd.user_id::text
		FROM product_variants v
		JOIN products p ON p.id = v.product_id
		JOIN vendors vd ON vd.id = p.vendor_id
		WHERE v.id = $1`, variantID).Scan(&ownerID, &productName, &sku, &vendorUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return InventoryRecord{}, errProductNotFound
	}
	if err != nil {
		return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
	}
	if ownerID != vendorID {
		return InventoryRecord{}, errProductNotFound
	}

	locationID, err := ensureLocation(ctx, tx, vendorID)
	if err != nil {
		return InventoryRecord{}, err
	}

	var current, reserved, reorder int
	err = tx.QueryRow(ctx, `
		SELECT quantity, reserved_quantity, reorder_level FROM inventory
		WHERE variant_id = $1 AND location_id = $2`, variantID, locationID).Scan(&current, &reserved, &reorder)
	if errors.Is(err, pgx.ErrNoRows) {
		current, reserved, reorder = 0, 0, 0
	} else if err != nil {
		return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
	}

	if quantity < reserved {
		return InventoryRecord{}, errStockBelowReserved
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO inventory (variant_id, location_id, quantity, reorder_level)
		VALUES ($1, $2, $3, 0)
		ON CONFLICT (variant_id, location_id)
		DO UPDATE SET quantity = EXCLUDED.quantity, updated_at = NOW()`,
		variantID, locationID, quantity); err != nil {
		return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
	}

	if delta := quantity - current; delta != 0 {
		movement := "restock"
		if delta < 0 {
			movement = "adjustment"
		}
		var notes *string
		if trimmed := strings.TrimSpace(note); trimmed != "" {
			notes = &trimmed
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO inventory_movements
				(variant_id, location_id, movement_type, quantity, quantity_before, quantity_after, notes, created_by)
			VALUES ($1, $2, $3::inventory_movement_type, $4, $5, $6, $7, $8)`,
			variantID, locationID, movement, absInt(delta), current, quantity, notes, userID); err != nil {
			return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
		}
	}

	// Alert the vendor when stock worsens into "low" or "out". This fires only
	// on a crossing, so an unchanged condition never re-notifies.
	if stockSeverity(quantity, reorder) > stockSeverity(current, reorder) {
		title, message := stockAlert(productName, sku, quantity, reorder)
		if err := notify.Insert(ctx, tx, vendorUserID, "system", title, message, map[string]any{
			"variant_id": variantID,
			"sku":        sku,
			"quantity":   quantity,
		}); err != nil {
			return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return r.inventoryByVariant(ctx, vendorID, variantID)
}

// stockSeverity ranks a stock level: 0 ok, 1 low, 2 out.
func stockSeverity(quantity, reorder int) int {
	switch {
	case quantity <= 0:
		return 2
	case reorder > 0 && quantity <= reorder:
		return 1
	default:
		return 0
	}
}

// stockAlert builds the notification title/message for a worsening stock level.
func stockAlert(name, sku string, quantity, reorder int) (string, string) {
	label := name
	if sku != "" {
		label = fmt.Sprintf("%s (%s)", name, sku)
	}
	if quantity <= 0 {
		return "Out of stock", fmt.Sprintf("%s is out of stock.", label)
	}
	return "Low stock", fmt.Sprintf("%s is down to %d units (threshold %d).", label, quantity, reorder)
}

// inventoryByVariant loads one variant's inventory aggregate for the vendor.
func (r *Repository) inventoryByVariant(ctx context.Context, vendorID, variantID string) (InventoryRecord, error) {
	var rec InventoryRecord
	err := r.db.QueryRow(ctx, `
		SELECT `+inventoryColumns+inventoryFrom+`
		WHERE v.id = $1 AND p.vendor_id = $2
		GROUP BY v.id, p.id, c.name, img.url`, variantID, vendorID).Scan(
		&rec.VariantID, &rec.ProductID, &rec.Name, &rec.SKU, &rec.Category,
		&rec.ImageURL, &rec.Quantity, &rec.Reserved, &rec.ReorderLevel,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return InventoryRecord{}, errProductNotFound
	}
	if err != nil {
		return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return rec, nil
}

// ListInventoryMovements returns one page of the vendor's stock movements,
// newest first, with the stock level before and after each change.
func (r *Repository) ListInventoryMovements(ctx context.Context, vendorID string, q ListQuery) ([]InventoryMovementRecord, int64, error) {
	args := []any{vendorID}
	where := " WHERE p.vendor_id = $1"
	if q.VariantID != "" {
		args = append(args, q.VariantID)
		where += " AND m.variant_id = $" + strconv.Itoa(len(args))
	}
	if q.Type != "" {
		args = append(args, q.Type)
		where += " AND m.movement_type = $" + strconv.Itoa(len(args)) + "::inventory_movement_type"
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		place := "$" + strconv.Itoa(len(args))
		where += " AND (p.name ILIKE " + place + " OR v.sku ILIKE " + place + ")"
	}

	const from = `
		FROM inventory_movements m
		JOIN product_variants v ON v.id = m.variant_id
		JOIN products p ON p.id = v.product_id
		LEFT JOIN orders o ON m.reference_type = 'order' AND o.id = m.reference_id
		LEFT JOIN users u ON u.id = m.created_by
		LEFT JOIN user_profiles up ON up.user_id = u.id`

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT count(*)`+from+where, args...).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []InventoryMovementRecord{}, 0, nil
	}

	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.db.Query(ctx, `
		SELECT m.id, v.id, p.id, p.name, v.sku, m.movement_type::text, m.quantity,
		       m.quantity_before, m.quantity_after, COALESCE(m.notes, ''),
		       COALESCE(m.reference_type, ''), COALESCE(o.order_number, ''),
		       COALESCE(NULLIF(TRIM(CONCAT_WS(' ', up.first_name, up.last_name)), ''),
		                split_part(COALESCE(u.email, ''), '@', 1), ''),
		       m.created_at`+from+where+`
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT $`+strconv.Itoa(len(args)+1)+` OFFSET $`+strconv.Itoa(len(args)+2), listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]InventoryMovementRecord, 0, q.Limit)
	for rows.Next() {
		var rec InventoryMovementRecord
		if err := rows.Scan(
			&rec.ID, &rec.VariantID, &rec.ProductID, &rec.Name, &rec.SKU, &rec.MovementType,
			&rec.Quantity, &rec.QuantityBefore, &rec.QuantityAfter, &rec.Notes,
			&rec.ReferenceType, &rec.OrderNumber, &rec.Actor, &rec.CreatedAt,
		); err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	return records, total, nil
}

// SetReorderLevel updates the low-stock threshold for a variant the vendor owns.
func (r *Repository) SetReorderLevel(ctx context.Context, vendorID, variantID, userID string, level int) (InventoryRecord, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var owned bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM product_variants v
			JOIN products p ON p.id = v.product_id
			WHERE v.id = $1 AND p.vendor_id = $2
		)`, variantID, vendorID).Scan(&owned); err != nil {
		return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
	}
	if !owned {
		return InventoryRecord{}, errProductNotFound
	}

	locationID, err := ensureLocation(ctx, tx, vendorID)
	if err != nil {
		return InventoryRecord{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO inventory (variant_id, location_id, quantity, reorder_level)
		VALUES ($1, $2, 0, $3)
		ON CONFLICT (variant_id, location_id)
		DO UPDATE SET reorder_level = EXCLUDED.reorder_level, updated_at = NOW()`,
		variantID, locationID, level); err != nil {
		return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id)
		VALUES ($1, 'inventory.threshold_changed', 'product_variant', $2)`,
		userID, variantID); err != nil {
		return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return InventoryRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return r.inventoryByVariant(ctx, vendorID, variantID)
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
