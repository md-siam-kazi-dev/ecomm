package vendor

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// pgUniqueViolation is PostgreSQL's unique_violation SQLSTATE.
const pgUniqueViolation = "23505"

// CreateProduct persists the product, its default variant, inventory, category
// link and images in a single transaction. Approval fields are set by the
// server and never taken from the payload.
func (r *Repository) CreateProduct(ctx context.Context, vendorID string, p ProductParams) (ProductDetail, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	brandID, err := upsertBrand(ctx, tx, p.Brand)
	if err != nil {
		return ProductDetail{}, err
	}

	slug, err := uniqueSlug(ctx, tx, slugify(p.Name))
	if err != nil {
		return ProductDetail{}, err
	}

	attributes := marshalAttributes(p.Attributes)

	var submittedAt *time.Time
	if p.Submit {
		now := time.Now()
		submittedAt = &now
	}

	var productID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO products
			(vendor_id, brand_id, name, slug, description, short_description, status, approval_status, attributes, submitted_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'draft', 'pending', $7::jsonb, $8)
		RETURNING id
	`, vendorID, brandID, p.Name, slug, p.Description, p.ShortDescription, attributes, submittedAt,
	).Scan(&productID); err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}

	if err := linkCategory(ctx, tx, productID, p.CategorySlug); err != nil {
		return ProductDetail{}, err
	}

	variantID, err := insertVariant(ctx, tx, productID, p)
	if err != nil {
		return ProductDetail{}, err
	}

	locationID, err := ensureLocation(ctx, tx, vendorID)
	if err != nil {
		return ProductDetail{}, err
	}

	if err := setStock(ctx, tx, variantID, locationID, p.Stock, p.LowStockThreshold); err != nil {
		return ProductDetail{}, err
	}

	if err := replaceImages(ctx, tx, productID, p.Images); err != nil {
		return ProductDetail{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}
	return r.GetProduct(ctx, vendorID, productID)
}

// UpdateProduct edits a product the vendor owns. Edited products return to
// pending review (drafts stay drafts until submitted).
func (r *Repository) UpdateProduct(ctx context.Context, vendorID, productID string, p ProductParams) (ProductDetail, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var owned bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1 AND vendor_id = $2)`, productID, vendorID).Scan(&owned); err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}
	if !owned {
		return ProductDetail{}, errProductNotFound
	}

	brandID, err := upsertBrand(ctx, tx, p.Brand)
	if err != nil {
		return ProductDetail{}, err
	}

	attributes := marshalAttributes(p.Attributes)
	submittedAt := "submitted_at"
	if p.Submit {
		submittedAt = "NOW()"
	}

	if _, err := tx.Exec(ctx, `
		UPDATE products
		SET name = $3, description = $4, short_description = $5, brand_id = $6,
		    attributes = $7::jsonb, status = 'draft', approval_status = 'pending',
		    review_note = NULL, submitted_at = `+submittedAt+`, updated_at = NOW()
		WHERE id = $1 AND vendor_id = $2
	`, productID, vendorID, p.Name, p.Description, p.ShortDescription, brandID, attributes); err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM product_categories WHERE product_id = $1`, productID); err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}
	if err := linkCategory(ctx, tx, productID, p.CategorySlug); err != nil {
		return ProductDetail{}, err
	}

	// Update (or create) the base variant.
	var variantID string
	err = tx.QueryRow(ctx, `SELECT id FROM product_variants WHERE product_id = $1 ORDER BY created_at LIMIT 1`, productID).Scan(&variantID)
	if errors.Is(err, pgx.ErrNoRows) {
		if variantID, err = insertVariant(ctx, tx, productID, p); err != nil {
			return ProductDetail{}, err
		}
	} else if err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	} else {
		sku := p.SKU
		if sku == "" {
			sku = generateSKU(p.Name)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE product_variants SET sku = $2, price = $3, compare_at_price = $4, updated_at = NOW()
			WHERE id = $1
		`, variantID, sku, p.Price, p.SalePrice); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
				return ProductDetail{}, errSKUTaken
			}
			return ProductDetail{}, apperr.ErrInternal.WithCause(err)
		}
	}

	locationID, err := ensureLocation(ctx, tx, vendorID)
	if err != nil {
		return ProductDetail{}, err
	}
	if err := setStock(ctx, tx, variantID, locationID, p.Stock, p.LowStockThreshold); err != nil {
		return ProductDetail{}, err
	}

	if err := replaceImages(ctx, tx, productID, p.Images); err != nil {
		return ProductDetail{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}
	return r.GetProduct(ctx, vendorID, productID)
}

// GetProduct returns a product the vendor owns, or errProductNotFound.
func (r *Repository) GetProduct(ctx context.Context, vendorID, productID string) (ProductDetail, error) {
	var (
		d         ProductDetail
		rawAttrs  []byte
		sku       string
		price     float64
		salePrice *float64
		variantID string
	)
	err := r.db.QueryRow(ctx, `
		SELECT p.id, p.name, p.slug, COALESCE(p.description, ''), COALESCE(p.short_description, ''),
		       p.status::text, p.approval_status::text,
		       COALESCE(c.name, ''), COALESCE(b.name, ''),
		       COALESCE(p.review_note, ''), p.submitted_at, p.approved_at, p.reviewed_at, p.created_at,
		       COALESCE(p.attributes, '{}'::jsonb)
		FROM products p
		LEFT JOIN brands b ON b.id = p.brand_id
		LEFT JOIN product_categories pc ON pc.product_id = p.id
		LEFT JOIN categories c ON c.id = pc.category_id
		WHERE p.id = $1 AND p.vendor_id = $2
		LIMIT 1
	`, productID, vendorID).Scan(
		&d.ID, &d.Name, &d.Slug, &d.Description, &d.ShortDescription,
		&d.Status, &d.ApprovalStatus, &d.Category, &d.Brand,
		&d.ReviewNote, &d.SubmittedAt, &d.ApprovedAt, &d.ReviewedAt, &d.CreatedAt, &rawAttrs,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProductDetail{}, errProductNotFound
	}
	if err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}

	d.Attributes = unmarshalAttributes(rawAttrs)

	// Base variant.
	if err := r.db.QueryRow(ctx, `
		SELECT id, sku, price, compare_at_price
		FROM product_variants WHERE product_id = $1 ORDER BY created_at LIMIT 1
	`, productID).Scan(&variantID, &sku, &price, &salePrice); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}
	d.SKU = sku
	d.Price = price
	d.SalePrice = salePrice

	// Stock across the vendor's locations.
	if err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(i.quantity), 0)::bigint, COALESCE(MAX(i.reorder_level), 0)
		FROM inventory i JOIN product_variants v ON v.id = i.variant_id
		WHERE v.product_id = $1
	`, productID).Scan(&d.Stock, &d.LowStock); err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}

	// Images.
	rows, err := r.db.Query(ctx, `
		SELECT url, COALESCE(alt_text, ''), sort_order, is_primary
		FROM product_images WHERE product_id = $1
		ORDER BY is_primary DESC, sort_order
	`, productID)
	if err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	for rows.Next() {
		var img ProductImage
		if err := rows.Scan(&img.URL, &img.AltText, &img.SortOrder, &img.IsPrimary); err != nil {
			return ProductDetail{}, apperr.ErrInternal.WithCause(err)
		}
		d.Images = append(d.Images, img)
	}
	if err := rows.Err(); err != nil {
		return ProductDetail{}, apperr.ErrInternal.WithCause(err)
	}

	return d, nil
}

// ---- helpers --------------------------------------------------------------

func upsertBrand(ctx context.Context, tx pgx.Tx, name string) (*string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	slug := slugify(name)

	var id string
	err := tx.QueryRow(ctx, `SELECT id FROM brands WHERE slug = $1`, slug).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := tx.QueryRow(ctx, `INSERT INTO brands (name, slug) VALUES ($1, $2) RETURNING id`, name, slug).Scan(&id); err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
	} else if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return &id, nil
}

func uniqueSlug(ctx context.Context, tx pgx.Tx, base string) (string, error) {
	if base == "" {
		base = "product"
	}
	slug := base
	for i := 0; i < 5; i++ {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1)`, slug).Scan(&exists); err != nil {
			return "", apperr.ErrInternal.WithCause(err)
		}
		if !exists {
			return slug, nil
		}
		slug = base + "-" + strings.ToLower(uuid.NewString()[:6])
	}
	return base + "-" + strings.ToLower(uuid.NewString()[:6]), nil
}

func linkCategory(ctx context.Context, tx pgx.Tx, productID, categorySlug string) error {
	tag, err := tx.Exec(ctx, `
		INSERT INTO product_categories (product_id, category_id)
		SELECT $1, c.id FROM categories c WHERE c.slug = $2
	`, productID, categorySlug)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrValidation
	}
	return nil
}

func insertVariant(ctx context.Context, tx pgx.Tx, productID string, p ProductParams) (string, error) {
	sku := strings.TrimSpace(p.SKU)
	if sku == "" {
		sku = generateSKU(p.Name)
	}
	var variantID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO product_variants (product_id, sku, price, compare_at_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, productID, sku, p.Price, p.SalePrice).Scan(&variantID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return "", errSKUTaken
		}
		return "", apperr.ErrInternal.WithCause(err)
	}
	return variantID, nil
}

func ensureLocation(ctx context.Context, tx pgx.Tx, vendorID string) (string, error) {
	var locationID string
	err := tx.QueryRow(ctx, `
		SELECT id FROM inventory_locations WHERE vendor_id = $1 AND is_active ORDER BY created_at LIMIT 1
	`, vendorID).Scan(&locationID)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := tx.QueryRow(ctx, `
			INSERT INTO inventory_locations (vendor_id, name) VALUES ($1, 'Main Warehouse') RETURNING id
		`, vendorID).Scan(&locationID); err != nil {
			return "", apperr.ErrInternal.WithCause(err)
		}
		return locationID, nil
	}
	if err != nil {
		return "", apperr.ErrInternal.WithCause(err)
	}
	return locationID, nil
}

func setStock(ctx context.Context, tx pgx.Tx, variantID, locationID string, quantity, reorderLevel int) error {
	var existing int
	err := tx.QueryRow(ctx, `
		SELECT count(*) FROM inventory WHERE variant_id = $1 AND location_id = $2
	`, variantID, locationID).Scan(&existing)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	if existing == 0 {
		if _, err := tx.Exec(ctx, `
			INSERT INTO inventory (variant_id, location_id, quantity, reorder_level)
			VALUES ($1, $2, $3, $4)
		`, variantID, locationID, quantity, reorderLevel); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
		if quantity > 0 {
			if _, err := tx.Exec(ctx, `
				INSERT INTO inventory_movements
					(variant_id, location_id, movement_type, quantity, quantity_before, quantity_after, notes)
				VALUES ($1, $2, 'restock', $3, 0, $3, 'Initial stock')
			`, variantID, locationID, quantity); err != nil {
				return apperr.ErrInternal.WithCause(err)
			}
		}
		return nil
	}

	if _, err := tx.Exec(ctx, `
		UPDATE inventory SET quantity = $3, reorder_level = $4, updated_at = NOW()
		WHERE variant_id = $1 AND location_id = $2
	`, variantID, locationID, quantity, reorderLevel); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

func replaceImages(ctx context.Context, tx pgx.Tx, productID string, urls []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM product_images WHERE product_id = $1`, productID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	for i, url := range urls {
		if _, err := tx.Exec(ctx, `
			INSERT INTO product_images (product_id, url, sort_order, is_primary)
			VALUES ($1, $2, $3, $4)
		`, productID, url, i, i == 0); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
	}
	return nil
}

func generateSKU(name string) string {
	base := strings.ToUpper(slugify(name))
	if base == "" {
		base = "PRODUCT"
	}
	if len(base) > 12 {
		base = base[:12]
	}
	return base + "-" + strings.ToUpper(uuid.NewString()[:6])
}

// slugify produces a URL-safe slug.
func slugify(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func marshalAttributes(attrs map[string]string) string {
	if len(attrs) == 0 {
		return "{}"
	}
	raw, err := json.Marshal(attrs)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func unmarshalAttributes(raw []byte) map[string]string {
	out := map[string]string{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return out
}
