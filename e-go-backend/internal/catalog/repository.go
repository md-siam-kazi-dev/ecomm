package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// CatalogRepository is the persistence boundary for the public catalog.
type CatalogRepository interface {
	ListProducts(ctx context.Context, q ListQuery, ids []string) ([]ProductRecord, int64, error)
	ProductBySlug(ctx context.Context, slug string) (ProductRecord, error)
	RelatedProducts(ctx context.Context, slug string, limit int) ([]ProductRecord, error)
	ListCategories(ctx context.Context) ([]CategoryRecord, error)
	StoreBySlug(ctx context.Context, slug string) (StoreRecord, error)
	StoreProducts(ctx context.Context, slug string, q ListQuery) ([]ProductRecord, int64, error)
	ProductReviews(ctx context.Context, slug string, page, limit int) ([]ReviewRecord, int64, error)
	ProductRating(ctx context.Context, slug string) (RatingSummaryRecord, error)
	ProductQuestions(ctx context.Context, slug string, page, limit int) ([]QuestionRecord, int64, error)
	TopReviews(ctx context.Context, limit int) ([]ReviewRecord, error)
	// SitemapProducts and SitemapStores are the keyset scans backing the public
	// sitemap endpoints. They return only indexable rows plus the total count of
	// indexable rows, ignoring the cursor (see sitemap_repository.go).
	SitemapProducts(ctx context.Context, cursor SitemapCursor, limit int) ([]SitemapRecord, int64, error)
	SitemapStores(ctx context.Context, cursor SitemapCursor, limit int) ([]SitemapRecord, int64, error)
}

// Repository persists catalog data in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

type scanner interface{ Scan(dest ...any) error }

// catalogFrom joins everything a storefront product needs. Only aggregates that
// cannot be a plain column (price, stock, rating, sales, image) are lateral.
const catalogFrom = `
	FROM products p
	LEFT JOIN brands b ON b.id = p.brand_id
	LEFT JOIN product_categories pc ON pc.product_id = p.id
	LEFT JOIN categories c ON c.id = pc.category_id
	LEFT JOIN stores shop ON shop.vendor_id = p.vendor_id AND shop.status = 'active'
	LEFT JOIN LATERAL (
		SELECT url, COALESCE(alt_text, '') AS alt_text FROM product_images pi
		WHERE pi.product_id = p.id ORDER BY pi.is_primary DESC, pi.sort_order LIMIT 1
	) img ON TRUE
	LEFT JOIN LATERAL (
		SELECT MIN(v.price)::float8 AS price, MIN(v.compare_at_price)::float8 AS compare_at, MIN(v.sku) AS sku
		FROM product_variants v WHERE v.product_id = p.id
	) pr ON TRUE
	LEFT JOIN LATERAL (
		SELECT COALESCE(SUM(i.quantity), 0)::bigint AS qty
		FROM inventory i JOIN product_variants v ON v.id = i.variant_id WHERE v.product_id = p.id
	) st ON TRUE
	LEFT JOIN LATERAL (
		SELECT AVG(rr.rating)::float8 AS rating, count(*)::bigint AS review_count
		FROM reviews rr WHERE rr.product_id = p.id AND rr.status = 'published'
	) rt ON TRUE
	LEFT JOIN LATERAL (
		SELECT COALESCE(SUM(oi.quantity), 0)::bigint AS units
		FROM order_items oi WHERE oi.product_id = p.id
	) sl ON TRUE`

const catalogColumns = `
	p.id, p.slug, p.name, COALESCE(p.short_description, ''), COALESCE(p.description, ''),
	COALESCE(pr.price, 0) AS price, pr.compare_at AS compare_at_price, COALESCE(pr.sku, ''),
	COALESCE(st.qty, 0) AS stock, COALESCE(img.url, ''), COALESCE(img.alt_text, ''),
	COALESCE(c.name, ''), COALESCE(c.slug, ''), COALESCE(b.name, ''),
	COALESCE(shop.name, ''), COALESCE(shop.slug, ''),
	COALESCE(rt.rating, 0) AS rating, COALESCE(rt.review_count, 0) AS review_count,
	COALESCE(sl.units, 0) AS sales_units,
	(p.created_at > NOW() - INTERVAL '21 days') AS is_new,
	COALESCE(p.attributes, '{}'::jsonb), COALESCE(p.seo_title, ''), COALESCE(p.seo_description, ''),
	p.created_at, p.updated_at`

const liveProducts = `p.status = 'active' AND p.approval_status = 'approved'`

// buildFilter returns the WHERE clause and positional args for the catalog list.
func buildFilter(q ListQuery, ids []string) (string, []any) {
	conditions := []string{liveProducts}
	args := []any{}

	if q.Category != "" {
		args = append(args, q.Category)
		conditions = append(conditions, "c.slug = $"+strconv.Itoa(len(args)))
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		place := "$" + strconv.Itoa(len(args))
		conditions = append(conditions, "(p.name ILIKE "+place+" OR p.description ILIKE "+place+" OR b.name ILIKE "+place+" OR c.name ILIKE "+place+")")
	}
	if q.MinPrice != nil {
		args = append(args, *q.MinPrice)
		conditions = append(conditions, "COALESCE(pr.price, 0) >= $"+strconv.Itoa(len(args)))
	}
	if q.MaxPrice != nil {
		args = append(args, *q.MaxPrice)
		conditions = append(conditions, "COALESCE(pr.price, 0) <= $"+strconv.Itoa(len(args)))
	}
	if q.MinRating != nil {
		args = append(args, *q.MinRating)
		conditions = append(conditions, "COALESCE(rt.rating, 0) >= $"+strconv.Itoa(len(args)))
	}
	if q.InStock {
		conditions = append(conditions, "COALESCE(st.qty, 0) > 0")
	}
	if q.OnSale {
		conditions = append(conditions, "pr.compare_at IS NOT NULL AND pr.compare_at > pr.price")
	}
	if len(ids) > 0 {
		args = append(args, strings.Join(ids, ","))
		conditions = append(conditions, "p.id::text = ANY(string_to_array($"+strconv.Itoa(len(args))+", ','))")
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func orderBy(sort string) string {
	switch sort {
	case "newest":
		return " ORDER BY p.created_at DESC, p.id"
	case "price-asc":
		return " ORDER BY price ASC, p.id"
	case "price-desc":
		return " ORDER BY price DESC, p.id"
	case "rating":
		return " ORDER BY rating DESC, review_count DESC, p.id"
	case "discount":
		return " ORDER BY (COALESCE(pr.compare_at, 0) - COALESCE(pr.price, 0)) DESC, p.id"
	case "featured":
		fallthrough
	default:
		return " ORDER BY (COALESCE(rt.rating, 0) >= 4.0) DESC, rating DESC, p.created_at DESC, p.id"
	}
}

// ListProducts returns one page of live products plus the total count.
func (r *Repository) ListProducts(ctx context.Context, q ListQuery, ids []string) ([]ProductRecord, int64, error) {
	where, args := buildFilter(q, ids)

	var total int64
	if err := r.db.QueryRow(ctx, "SELECT count(*) "+catalogFrom+" "+where, args...).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []ProductRecord{}, 0, nil
	}

	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)
	query := "SELECT " + catalogColumns + catalogFrom + " " + where + orderBy(q.Sort) +
		" LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]ProductRecord, 0, q.Limit)
	for rows.Next() {
		rec, err := scanProduct(rows)
		if err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	return records, total, nil
}

// ProductBySlug returns one live product.
func (r *Repository) ProductBySlug(ctx context.Context, slug string) (ProductRecord, error) {
	rec, err := scanProduct(r.db.QueryRow(ctx, "SELECT "+catalogColumns+catalogFrom+" WHERE p.slug = $1 AND "+liveProducts, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return ProductRecord{}, errProductNotFound
	}
	if err != nil {
		return ProductRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return rec, nil
}

// RelatedProducts returns other live products in the same category, best first.
func (r *Repository) RelatedProducts(ctx context.Context, slug string, limit int) ([]ProductRecord, error) {
	rows, err := r.db.Query(ctx, "SELECT "+catalogColumns+catalogFrom+`
		WHERE `+liveProducts+` AND p.slug <> $1
		  AND pc.category_id = (
			SELECT pc2.category_id FROM products p2
			JOIN product_categories pc2 ON pc2.product_id = p2.id
			WHERE p2.slug = $1 LIMIT 1
		  )
		ORDER BY rating DESC, review_count DESC, p.created_at DESC, p.id
		LIMIT $2`, slug, limit)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	return collectProducts(rows, limit)
}

// ListCategories returns active categories with their live product counts.
func (r *Repository) ListCategories(ctx context.Context) ([]CategoryRecord, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.slug, c.name, COALESCE(c.description, ''), COALESCE(c.image_url, ''),
		       count(p.id)::bigint, c.updated_at
		FROM categories c
		LEFT JOIN product_categories pc ON pc.category_id = c.id
		LEFT JOIN products p ON p.id = pc.product_id AND `+liveProducts+`
		WHERE c.status = TRUE
		GROUP BY c.id, c.sort_order
		ORDER BY c.sort_order, c.name`)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]CategoryRecord, 0)
	for rows.Next() {
		var rec CategoryRecord
		if err := rows.Scan(&rec.ID, &rec.Slug, &rec.Name, &rec.Description, &rec.ImageURL, &rec.ProductCount, &rec.UpdatedAt); err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return records, nil
}

// StoreBySlug returns one active store with its public aggregates. Positive
// feedback is the share of published reviews rated 4 or 5 stars.
func (r *Repository) StoreBySlug(ctx context.Context, slug string) (StoreRecord, error) {
	var rec StoreRecord
	err := r.db.QueryRow(ctx, `
		SELECT s.id, s.name, s.slug, COALESCE(s.description, ''),
		       COALESCE(s.logo_url, ''), COALESCE(s.banner_url, ''), COALESCE(c.name, ''),
		       (SELECT count(*) FROM products p
		        WHERE p.vendor_id = s.vendor_id AND `+liveProducts+`)::bigint,
		       (SELECT count(*) FROM store_followers f WHERE f.store_id = s.id)::bigint,
		       (SELECT count(*) FROM reviews r
		        JOIN products p2 ON p2.id = r.product_id
		        WHERE p2.vendor_id = s.vendor_id AND r.status = 'published')::bigint,
		       (SELECT count(*) FROM reviews r
		        JOIN products p3 ON p3.id = r.product_id
		        WHERE p3.vendor_id = s.vendor_id AND r.status = 'published' AND r.rating >= 4)::bigint,
		       s.updated_at
		FROM stores s
		LEFT JOIN categories c ON c.id = s.category_id
		WHERE s.slug = $1 AND s.status = 'active'`, slug,
	).Scan(&rec.ID, &rec.Name, &rec.Slug, &rec.Description, &rec.LogoURL, &rec.BannerURL,
		&rec.Category, &rec.ProductCount, &rec.FollowerCount, &rec.ReviewCount, &rec.PositiveCount,
		&rec.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return StoreRecord{}, errStoreNotFound
	}
	if err != nil {
		return StoreRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return rec, nil
}

// StoreProducts returns one page of an active store's live products.
func (r *Repository) StoreProducts(ctx context.Context, slug string, q ListQuery) ([]ProductRecord, int64, error) {
	where := " WHERE shop.slug = $1 AND " + liveProducts

	var total int64
	if err := r.db.QueryRow(ctx, "SELECT count(*) "+catalogFrom+where, slug).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []ProductRecord{}, 0, nil
	}

	rows, err := r.db.Query(ctx, "SELECT "+catalogColumns+catalogFrom+where+orderBy(q.Sort)+
		" LIMIT $2 OFFSET $3", slug, q.Limit, (q.Page-1)*q.Limit)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records, err := collectProducts(rows, q.Limit)
	if err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

const reviewColumns = `
	r.id, r.product_id, p.name, p.slug,
	COALESCE(NULLIF(TRIM(CONCAT_WS(' ', up.first_name, up.last_name)), ''), split_part(u.email, '@', 1)),
	COALESCE(up.avatar_url, ''),
	r.rating, COALESCE(r.title, ''), r.content,
	(r.order_item_id IS NOT NULL) AS verified,
	COALESCE(v.helpful, 0)::bigint, COALESCE(r.vendor_reply, ''), r.vendor_replied_at, r.created_at`

const reviewFrom = `
	FROM reviews r
	JOIN products p ON p.id = r.product_id
	JOIN users u ON u.id = r.user_id
	LEFT JOIN user_profiles up ON up.user_id = r.user_id
	LEFT JOIN LATERAL (
		SELECT count(*) FILTER (WHERE rv.is_helpful)::bigint AS helpful
		FROM review_votes rv WHERE rv.review_id = r.id
	) v ON TRUE`

// ProductReviews returns one page of the published reviews for a product,
// newest first, plus the total count.
func (r *Repository) ProductReviews(ctx context.Context, slug string, page, limit int) ([]ReviewRecord, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `
		SELECT count(*)
		FROM reviews r
		JOIN products p ON p.id = r.product_id
		WHERE (p.slug = $1 OR p.id::text = $1) AND r.status = 'published'`, slug,
	).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []ReviewRecord{}, 0, nil
	}

	rows, err := r.db.Query(ctx, "SELECT "+reviewColumns+reviewFrom+`
		WHERE (p.slug = $1 OR p.id::text = $1) AND r.status = 'published'
		ORDER BY r.created_at DESC, r.id
		LIMIT $2 OFFSET $3`, slug, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records, err := collectReviews(rows, limit)
	if err != nil {
		return nil, 0, err
	}
	if err := r.attachReviewImages(ctx, records); err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

// attachReviewImages groups each review's photo URLs onto the records.
func (r *Repository) attachReviewImages(ctx context.Context, records []ReviewRecord) error {
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

// ProductRating aggregates a product's published reviews into a 1..5 star
// distribution and total count. The average is derived by the service.
func (r *Repository) ProductRating(ctx context.Context, slug string) (RatingSummaryRecord, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.rating, count(*)::bigint
		FROM reviews r
		JOIN products p ON p.id = r.product_id
		WHERE (p.slug = $1 OR p.id::text = $1) AND r.status = 'published'
		GROUP BY r.rating`, slug)
	if err != nil {
		return RatingSummaryRecord{}, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	var summary RatingSummaryRecord
	for rows.Next() {
		var (
			rating int
			count  int64
		)
		if err := rows.Scan(&rating, &count); err != nil {
			return RatingSummaryRecord{}, apperr.ErrInternal.WithCause(err)
		}
		if rating >= 1 && rating <= 5 {
			summary.Distribution[rating-1] = count
			summary.Total += count
		}
	}
	if err := rows.Err(); err != nil {
		return RatingSummaryRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return summary, nil
}

// ProductQuestions returns one page of the published questions for a product,
// newest first, each with its public answers, plus the total count.
func (r *Repository) ProductQuestions(ctx context.Context, slug string, page, limit int) ([]QuestionRecord, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `
		SELECT count(*)
		FROM product_questions q
		JOIN products p ON p.id = q.product_id
		WHERE (p.slug = $1 OR p.id::text = $1) AND q.is_published`, slug,
	).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []QuestionRecord{}, 0, nil
	}

	rows, err := r.db.Query(ctx, `
		SELECT q.id, q.question,
		       COALESCE(NULLIF(TRIM(CONCAT_WS(' ', up.first_name, up.last_name)), ''), split_part(u.email, '@', 1)),
		       q.created_at
		FROM product_questions q
		JOIN products p ON p.id = q.product_id
		JOIN users u ON u.id = q.user_id
		LEFT JOIN user_profiles up ON up.user_id = q.user_id
		WHERE (p.slug = $1 OR p.id::text = $1) AND q.is_published
		ORDER BY q.created_at DESC, q.id
		LIMIT $2 OFFSET $3`, slug, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]QuestionRecord, 0, limit)
	for rows.Next() {
		var q QuestionRecord
		if err := rows.Scan(&q.ID, &q.Question, &q.Author, &q.CreatedAt); err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		q.Answers = []AnswerRecord{}
		records = append(records, q)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if len(records) == 0 {
		return records, total, nil
	}

	ids := make([]string, 0, len(records))
	for _, rec := range records {
		ids = append(ids, rec.ID)
	}
	answers, err := r.answersForQuestions(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	for i := range records {
		if list, ok := answers[records[i].ID]; ok {
			records[i].Answers = list
		}
	}
	return records, total, nil
}

// answersForQuestions loads the public answers for a set of question ids.
func (r *Repository) answersForQuestions(ctx context.Context, ids []string) (map[string][]AnswerRecord, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.question_id, a.id, a.answer, a.is_official_vendor_answer,
		       COALESCE(NULLIF(TRIM(CONCAT_WS(' ', up.first_name, up.last_name)), ''), split_part(u.email, '@', 1)),
		       a.created_at
		FROM product_question_answers a
		JOIN users u ON u.id = a.user_id
		LEFT JOIN user_profiles up ON up.user_id = a.user_id
		WHERE a.question_id::text = ANY(string_to_array($1, ','))
		ORDER BY a.created_at, a.id`, strings.Join(ids, ","))
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	grouped := make(map[string][]AnswerRecord, len(ids))
	for rows.Next() {
		var (
			questionID string
			a          AnswerRecord
		)
		if err := rows.Scan(&questionID, &a.ID, &a.Answer, &a.IsOfficialVendorAnswer, &a.Author, &a.CreatedAt); err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		grouped[questionID] = append(grouped[questionID], a)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return grouped, nil
}

// TopReviews returns the highest-rated published reviews (home testimonials).
func (r *Repository) TopReviews(ctx context.Context, limit int) ([]ReviewRecord, error) {
	rows, err := r.db.Query(ctx, "SELECT "+reviewColumns+reviewFrom+`
		WHERE r.status = 'published' AND r.rating = 5
		ORDER BY v.helpful DESC, r.created_at DESC, r.id
		LIMIT $1`, limit)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records, err := collectReviews(rows, limit)
	if err != nil {
		return nil, err
	}
	if err := r.attachReviewImages(ctx, records); err != nil {
		return nil, err
	}
	return records, nil
}

func scanProduct(s scanner) (ProductRecord, error) {
	var (
		rec      ProductRecord
		rawAttrs []byte
	)
	err := s.Scan(
		&rec.ID, &rec.Slug, &rec.Name, &rec.ShortDescription, &rec.Description,
		&rec.Price, &rec.CompareAtPrice, &rec.SKU, &rec.Stock, &rec.ImageURL, &rec.ImageAlt,
		&rec.Category, &rec.CategorySlug, &rec.Brand,
		&rec.StoreName, &rec.StoreSlug,
		&rec.Rating, &rec.ReviewCount, &rec.SalesUnits, &rec.IsNew,
		&rawAttrs, &rec.SeoTitle, &rec.SeoDescription, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		return rec, err
	}
	rec.Attributes = map[string]string{}
	if len(rawAttrs) > 0 {
		_ = json.Unmarshal(rawAttrs, &rec.Attributes)
	}
	return rec, nil
}

func collectProducts(rows pgx.Rows, limit int) ([]ProductRecord, error) {
	records := make([]ProductRecord, 0, limit)
	for rows.Next() {
		rec, err := scanProduct(rows)
		if err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return records, nil
}

func collectReviews(rows pgx.Rows, limit int) ([]ReviewRecord, error) {
	records := make([]ReviewRecord, 0, limit)
	for rows.Next() {
		var rec ReviewRecord
		if err := rows.Scan(
			&rec.ID, &rec.ProductID, &rec.ProductName, &rec.ProductSlug, &rec.Author, &rec.AvatarURL,
			&rec.Rating, &rec.Title, &rec.Comment, &rec.VerifiedPurchase, &rec.HelpfulCount,
			&rec.VendorReply, &rec.VendorRepliedAt, &rec.CreatedAt,
		); err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return records, nil
}
