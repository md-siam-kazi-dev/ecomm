package vendor

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// rowScanner is shared by the review/question scan helpers.
type rowScanner interface{ Scan(dest ...any) error }

const vendorReviewFrom = `
	FROM reviews r
	JOIN products p ON p.id = r.product_id AND p.vendor_id = $1
	JOIN order_items oi ON oi.id = r.order_item_id
	JOIN orders o ON o.id = oi.order_id
	JOIN users u ON u.id = r.user_id
	LEFT JOIN user_profiles up ON up.user_id = r.user_id
	LEFT JOIN LATERAL (
		SELECT url FROM product_images pi
		WHERE pi.product_id = p.id
		ORDER BY pi.is_primary DESC, pi.sort_order
		LIMIT 1
	) img ON TRUE`

const vendorReviewColumns = `
	r.id, r.product_id, p.name, p.slug, COALESCE(img.url, ''),
	COALESCE(NULLIF(TRIM(CONCAT_WS(' ', up.first_name, up.last_name)), ''), split_part(u.email, '@', 1)),
	o.order_number, r.rating, COALESCE(r.title, ''), r.content, r.created_at,
	COALESCE(r.vendor_reply, ''), r.vendor_replied_at`

func scanVendorReview(s rowScanner) (VendorReviewRecord, error) {
	var rec VendorReviewRecord
	err := s.Scan(&rec.ID, &rec.ProductID, &rec.ProductName, &rec.ProductSlug, &rec.ProductImage,
		&rec.Author, &rec.OrderNumber, &rec.Rating, &rec.Title, &rec.Comment, &rec.CreatedAt,
		&rec.VendorReply, &rec.VendorRepliedAt)
	return rec, err
}

// ListReviews returns one page of reviews on the vendor's products plus the
// total matching count.
func (r *Repository) ListReviews(ctx context.Context, vendorID string, q ListQuery) ([]VendorReviewRecord, int64, error) {
	args := []any{vendorID}
	where := ""
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += " AND p.name ILIKE $" + strconv.Itoa(len(args))
	}
	switch q.Status {
	case "replied":
		where += " AND r.vendor_reply IS NOT NULL AND r.vendor_reply <> ''"
	case "unreplied":
		where += " AND (r.vendor_reply IS NULL OR r.vendor_reply = '')"
	}

	var total int64
	if err := r.db.QueryRow(ctx, "SELECT count(*)"+vendorReviewFrom+where, args...).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []VendorReviewRecord{}, 0, nil
	}

	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.db.Query(ctx, "SELECT "+vendorReviewColumns+vendorReviewFrom+where+`
		ORDER BY r.created_at DESC, r.id
		LIMIT $`+strconv.Itoa(len(args)+1)+` OFFSET $`+strconv.Itoa(len(args)+2), listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]VendorReviewRecord, 0, q.Limit)
	for rows.Next() {
		rec, err := scanVendorReview(rows)
		if err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if err := r.attachReviewImages(ctx, records); err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

// attachReviewImages groups each review's photo URLs onto the records.
func (r *Repository) attachReviewImages(ctx context.Context, records []VendorReviewRecord) error {
	if len(records) == 0 {
		return nil
	}
	ids := make([]string, 0, len(records))
	for _, rec := range records {
		ids = append(ids, rec.ID)
	}
	rows, err := r.db.Query(ctx, `
		SELECT review_id, image_url FROM review_images
		WHERE review_id::text = ANY(string_to_array($1, ','))
		ORDER BY review_id, sort_order`, strings.Join(ids, ","))
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	byReview := make(map[string][]string, len(ids))
	for rows.Next() {
		var (
			reviewID string
			url      string
		)
		if err := rows.Scan(&reviewID, &url); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
		byReview[reviewID] = append(byReview[reviewID], url)
	}
	if err := rows.Err(); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	for i := range records {
		if images, ok := byReview[records[i].ID]; ok {
			records[i].Images = images
		}
	}
	return nil
}

// ReviewSummary aggregates the vendor's star feedback and reply backlog.
func (r *Repository) ReviewSummary(ctx context.Context, vendorID string) (ReviewSummaryRecord, error) {
	var (
		avg       float64
		total     int64
		d1, d2    int64
		d3, d4    int64
		d5        int64
		unreplied int64
	)
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(r.rating), 0)::float8, count(*)::bigint,
		       count(*) FILTER (WHERE r.rating = 1)::bigint,
		       count(*) FILTER (WHERE r.rating = 2)::bigint,
		       count(*) FILTER (WHERE r.rating = 3)::bigint,
		       count(*) FILTER (WHERE r.rating = 4)::bigint,
		       count(*) FILTER (WHERE r.rating = 5)::bigint,
		       count(*) FILTER (WHERE r.vendor_reply IS NULL OR r.vendor_reply = '')::bigint
		FROM reviews r
		JOIN products p ON p.id = r.product_id AND p.vendor_id = $1`, vendorID,
	).Scan(&avg, &total, &d1, &d2, &d3, &d4, &d5, &unreplied)
	if err != nil {
		return ReviewSummaryRecord{}, apperr.ErrInternal.WithCause(err)
	}

	summary := ReviewSummaryRecord{Total: total, Unreplied: unreplied}
	summary.Distribution = [5]int64{d1, d2, d3, d4, d5}
	return summary, nil
}

// ReplyToReview sets (or replaces) the vendor's reply on one of their reviews.
func (r *Repository) ReplyToReview(ctx context.Context, vendorID, reviewID, reply string) (VendorReviewRecord, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		UPDATE reviews
		SET vendor_reply = $3, vendor_replied_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND product_id IN (SELECT id FROM products WHERE vendor_id = $2)
		RETURNING id`, reviewID, vendorID, reply).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return VendorReviewRecord{}, errReviewNotFound
	}
	if err != nil {
		return VendorReviewRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return r.reviewByID(ctx, vendorID, reviewID)
}

// reviewByID loads one review (plus its photos) scoped to the vendor's products.
func (r *Repository) reviewByID(ctx context.Context, vendorID, reviewID string) (VendorReviewRecord, error) {
	rec, err := scanVendorReview(r.db.QueryRow(ctx,
		"SELECT "+vendorReviewColumns+vendorReviewFrom+" WHERE r.id = $2", vendorID, reviewID))
	if errors.Is(err, pgx.ErrNoRows) {
		return VendorReviewRecord{}, errReviewNotFound
	}
	if err != nil {
		return VendorReviewRecord{}, apperr.ErrInternal.WithCause(err)
	}
	records := []VendorReviewRecord{rec}
	if err := r.attachReviewImages(ctx, records); err != nil {
		return VendorReviewRecord{}, err
	}
	return records[0], nil
}
