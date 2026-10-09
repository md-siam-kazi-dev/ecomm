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
	ProductReviews(ctx context.Context, slug string, limit int) ([]ReviewRecord, error)
	TopReviews(ctx context.Context, limit int) ([]ReviewRecord, error)
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
		SELECT c.id, c.slug, c.name, COALESCE(c.description, ''), COALESCE(c.image_url, ''), count(p.id)::bigint
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
		if err := rows.Scan(&rec.ID, &rec.Slug, &rec.Name, &rec.Description, &rec.ImageURL, &rec.ProductCount); err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return records, nil
}

const reviewColumns = `
	r.id, r.product_id, p.name, p.slug,
	COALESCE(NULLIF(TRIM(CONCAT_WS(' ', up.first_name, up.last_name)), ''), split_part(u.email, '@', 1)),
	r.rating, COALESCE(r.title, ''), r.content,
	(r.order_item_id IS NOT NULL) AS verified,
	COALESCE(v.helpful, 0)::bigint, r.created_at`

const reviewFrom = `
	FROM reviews r
	JOIN products p ON p.id = r.product_id
	JOIN users u ON u.id = r.user_id
	LEFT JOIN user_profiles up ON up.user_id = r.user_id
	LEFT JOIN LATERAL (
		SELECT count(*) FILTER (WHERE rv.is_helpful)::bigint AS helpful
		FROM review_votes rv WHERE rv.review_id = r.id
	) v ON TRUE`

// ProductReviews returns the published reviews for a product, newest first.
func (r *Repository) ProductReviews(ctx context.Context, slug string, limit int) ([]ReviewRecord, error) {
	rows, err := r.db.Query(ctx, "SELECT "+reviewColumns+reviewFrom+`
		WHERE (p.slug = $1 OR p.id::text = $1) AND r.status = 'published'
		ORDER BY r.created_at DESC, r.id
		LIMIT $2`, slug, limit)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	return collectReviews(rows, limit)
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
	return collectReviews(rows, limit)
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
			&rec.ID, &rec.ProductID, &rec.ProductName, &rec.ProductSlug, &rec.Author,
			&rec.Rating, &rec.Title, &rec.Comment, &rec.VerifiedPurchase, &rec.HelpfulCount, &rec.CreatedAt,
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
