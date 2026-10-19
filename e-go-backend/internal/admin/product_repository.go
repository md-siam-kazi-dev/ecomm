package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/md-siam-kazi/e-go-backend/internal/audit"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// errProductNotFound is returned when the product does not exist.
var errProductNotFound = apperr.New(http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")

// ProductRepository is the persistence boundary for product moderation.
type ProductRepository interface {
	ListForReview(ctx context.Context, q ReviewQuery) ([]ReviewProduct, int64, error)
	GetForReview(ctx context.Context, productID string) (ReviewProductDetail, error)
	SetApproval(ctx context.Context, productID, adminID string, decision Decision, note string) error
}

// ListForReview returns submitted products awaiting / needing moderation.
// Only products that have been submitted (submitted_at set) are included.
func (r *Repository) ListForReview(ctx context.Context, q ReviewQuery) ([]ReviewProduct, int64, error) {
	args := []any{}
	status := ""
	switch q.Status {
	case "pending":
		status = " AND p.approval_status = 'pending'"
	case "rejected":
		status = " AND p.approval_status = 'rejected'"
	case "changes_requested":
		status = " AND p.approval_status = 'changes_requested'"
	}

	search := ""
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		place := "$" + strconv.Itoa(len(args))
		search = " AND p.name ILIKE " + place
	}

	baseWhere := ` WHERE p.submitted_at IS NOT NULL` + status + search

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM products p`+baseWhere, args...).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []ReviewProduct{}, 0, nil
	}

	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.name, COALESCE(v.business_name, ''), p.status::text, p.approval_status::text,
		       COALESCE((SELECT pv.price FROM product_variants pv WHERE pv.product_id = p.id ORDER BY pv.created_at LIMIT 1), 0)::float8,
		       COALESCE((SELECT SUM(i.quantity) FROM inventory i JOIN product_variants pv ON pv.id = i.variant_id WHERE pv.product_id = p.id), 0)::bigint,
		       COALESCE((SELECT img.url FROM product_images img WHERE img.product_id = p.id ORDER BY img.is_primary DESC, img.sort_order LIMIT 1), ''),
		       COALESCE(p.review_note, ''), p.submitted_at, p.reviewed_at, p.created_at
		FROM products p
		LEFT JOIN vendors v ON v.id = p.vendor_id
	`+baseWhere+`
		ORDER BY p.submitted_at DESC, p.created_at DESC
		LIMIT $`+strconv.Itoa(len(args)+1)+` OFFSET $`+strconv.Itoa(len(args)+2),
		listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]ReviewProduct, 0, q.Limit)
	for rows.Next() {
		var rec ReviewProduct
		if err := rows.Scan(
			&rec.ID, &rec.Name, &rec.VendorName, &rec.Status, &rec.ApprovalStatus,
			&rec.Price, &rec.Stock, &rec.ImageURL, &rec.ReviewNote, &rec.SubmittedAt, &rec.ReviewedAt, &rec.CreatedAt,
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

// GetForReview loads one submitted product with its vendor.
func (r *Repository) GetForReview(ctx context.Context, productID string) (ReviewProductDetail, error) {
	var (
		d        ReviewProductDetail
		rawAttrs []byte
		sku      string
		price    float64
		sale     *float64
	)
	err := r.db.QueryRow(ctx, `
		SELECT p.id, p.name, p.slug, COALESCE(p.description, ''), COALESCE(p.short_description, ''),
		       p.status::text, p.approval_status::text,
		       COALESCE(c.name, ''), COALESCE(b.name, ''),
		       COALESCE(p.review_note, ''), COALESCE(v.business_name, ''),
		       p.submitted_at, p.approved_at, p.reviewed_at, p.created_at,
		       COALESCE(p.attributes, '{}'::jsonb)
		FROM products p
		LEFT JOIN vendors v ON v.id = p.vendor_id
		LEFT JOIN brands b ON b.id = p.brand_id
		LEFT JOIN product_categories pc ON pc.product_id = p.id
		LEFT JOIN categories c ON c.id = pc.category_id
		WHERE p.id = $1
		LIMIT 1
	`, productID).Scan(
		&d.ID, &d.Name, &d.Slug, &d.Description, &d.ShortDescription,
		&d.Status, &d.ApprovalStatus, &d.Category, &d.Brand,
		&d.ReviewNote, &d.VendorName, &d.SubmittedAt, &d.ApprovedAt, &d.ReviewedAt, &d.CreatedAt, &rawAttrs,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ReviewProductDetail{}, errProductNotFound
	}
	if err != nil {
		return ReviewProductDetail{}, apperr.ErrInternal.WithCause(err)
	}

	d.Attributes = map[string]string{}
	if len(rawAttrs) > 0 {
		_ = json.Unmarshal(rawAttrs, &d.Attributes)
	}

	if err := r.db.QueryRow(ctx, `
		SELECT sku, price, compare_at_price FROM product_variants WHERE product_id = $1 ORDER BY created_at LIMIT 1
	`, productID).Scan(&sku, &price, &sale); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ReviewProductDetail{}, apperr.ErrInternal.WithCause(err)
	}
	d.SKU = sku
	d.Price = price
	d.SalePrice = sale

	if err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(i.quantity), 0)::bigint, COALESCE(MAX(i.reorder_level), 0)
		FROM inventory i JOIN product_variants pv ON pv.id = i.variant_id WHERE pv.product_id = $1
	`, productID).Scan(&d.Stock, &d.LowStock); err != nil {
		return ReviewProductDetail{}, apperr.ErrInternal.WithCause(err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT url, COALESCE(alt_text, ''), sort_order, is_primary
		FROM product_images WHERE product_id = $1
		ORDER BY is_primary DESC, sort_order
	`, productID)
	if err != nil {
		return ReviewProductDetail{}, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	for rows.Next() {
		var img ReviewImage
		if err := rows.Scan(&img.URL, &img.Alt, &img.SortOrder, &img.IsPrimary); err != nil {
			return ReviewProductDetail{}, apperr.ErrInternal.WithCause(err)
		}
		d.Images = append(d.Images, img)
	}
	if err := rows.Err(); err != nil {
		return ReviewProductDetail{}, apperr.ErrInternal.WithCause(err)
	}

	return d, nil
}

// SetApproval applies an admin decision. Approve publishes (eligible); reject
// and request-changes keep the product as a draft and store the note. The state
// change and its audit row are committed in a single transaction so a product
// can never change without a matching audit entry.
func (r *Repository) SetApproval(ctx context.Context, productID, adminID string, decision Decision, note string) error {
	action, newStatus, newApproval, err := approvalOutcome(decision)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var oldStatus, oldApproval string
	err = tx.QueryRow(ctx, `
		SELECT status::text, approval_status::text FROM products WHERE id = $1 FOR UPDATE
	`, productID).Scan(&oldStatus, &oldApproval)
	if errors.Is(err, pgx.ErrNoRows) {
		return errProductNotFound
	}
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	var tag pgconn.CommandTag
	reviewNote := any(nil)
	switch decision {
	case DecisionApprove:
		tag, err = tx.Exec(ctx, `
			UPDATE products
			SET status = 'active', approval_status = 'approved',
			    approved_by = $2, approved_at = NOW(), reviewed_by = $2, reviewed_at = NOW(),
			    review_note = NULL, updated_at = NOW()
			WHERE id = $1
		`, productID, adminID)
	default:
		reviewNote = note
		tag, err = tx.Exec(ctx, `
			UPDATE products
			SET status = 'draft', approval_status = $2,
			    approved_by = NULL, approved_at = NULL, reviewed_by = $3, reviewed_at = NOW(),
			    review_note = $4, updated_at = NOW()
			WHERE id = $1
		`, productID, newApproval, adminID, note)
	}
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return errProductNotFound
	}

	actorID := adminID
	entityID := productID
	if err := audit.Record(ctx, tx, audit.Event{
		ActorID:    &actorID,
		ActorRole:  audit.RoleAdmin,
		Action:     action,
		EntityType: audit.EntityProduct,
		EntityID:   &entityID,
		Result:     audit.ResultSuccess,
		Reason:     note,
		OldValues:  map[string]any{"status": oldStatus, "approval_status": oldApproval},
		NewValues: map[string]any{
			"status":          newStatus,
			"approval_status": newApproval,
			"review_note":     reviewNote,
		},
	}); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// approvalOutcome maps a decision to its audit action and resulting state.
func approvalOutcome(decision Decision) (action, status, approval string, err error) {
	switch decision {
	case DecisionApprove:
		return audit.ActionAdminProductApproved, "active", "approved", nil
	case DecisionReject:
		return audit.ActionAdminProductRejected, "draft", "rejected", nil
	case DecisionChanges:
		return audit.ActionAdminProductChangesRequested, "draft", "changes_requested", nil
	default:
		return "", "", "", apperr.ErrValidation
	}
}
