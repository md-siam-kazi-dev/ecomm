package catalog

import (
	"context"
	"sort"
	"time"
)

// fakeRepo is a configurable in-memory CatalogRepository.
type fakeRepo struct {
	products  []ProductRecord
	total     int64
	lastQuery ListQuery
	lastIDs   []string

	product    ProductRecord
	productErr error

	related            []ProductRecord
	categories         []CategoryRecord
	store              StoreRecord
	storeErr           error
	storeProducts      []ProductRecord
	storeProductsTotal int64
	lastStoreQuery     ListQuery
	reviews            []ReviewRecord
	reviewsTotal       int64
	rating             RatingSummaryRecord
	questions          []QuestionRecord
	questionsTotal     int64
	top                []ReviewRecord

	// Sitemap scans. The fake applies the same visibility predicate as the SQL
	// so the endpoints can be exercised without a database.
	sitemapProducts   []sitemapFixture
	sitemapStores     []sitemapFixture
	lastSitemapCursor SitemapCursor
	lastSitemapLimit  int
}

func (f *fakeRepo) ListProducts(_ context.Context, q ListQuery, ids []string) ([]ProductRecord, int64, error) {
	f.lastQuery, f.lastIDs = q, ids
	return f.products, f.total, nil
}

func (f *fakeRepo) ProductBySlug(context.Context, string) (ProductRecord, error) {
	if f.productErr != nil {
		return ProductRecord{}, f.productErr
	}
	return f.product, nil
}

func (f *fakeRepo) RelatedProducts(context.Context, string, int) ([]ProductRecord, error) {
	return f.related, nil
}

func (f *fakeRepo) ListCategories(context.Context) ([]CategoryRecord, error) {
	return f.categories, nil
}

func (f *fakeRepo) StoreBySlug(context.Context, string) (StoreRecord, error) {
	if f.storeErr != nil {
		return StoreRecord{}, f.storeErr
	}
	return f.store, nil
}

func (f *fakeRepo) StoreProducts(_ context.Context, _ string, q ListQuery) ([]ProductRecord, int64, error) {
	f.lastStoreQuery = q
	return f.storeProducts, f.storeProductsTotal, nil
}

func (f *fakeRepo) ProductReviews(context.Context, string, int, int) ([]ReviewRecord, int64, error) {
	return f.reviews, f.reviewsTotal, nil
}

func (f *fakeRepo) ProductRating(context.Context, string) (RatingSummaryRecord, error) {
	return f.rating, nil
}

func (f *fakeRepo) ProductQuestions(context.Context, string, int, int) ([]QuestionRecord, int64, error) {
	return f.questions, f.questionsTotal, nil
}

func (f *fakeRepo) TopReviews(context.Context, int) ([]ReviewRecord, error) {
	return f.top, nil
}

func (f *fakeRepo) SitemapProducts(_ context.Context, cursor SitemapCursor, limit int) ([]SitemapRecord, int64, error) {
	f.lastSitemapCursor, f.lastSitemapLimit = cursor, limit
	records, total := productSitemapPage(f.sitemapProducts, cursor, limit)
	return records, total, nil
}

func (f *fakeRepo) SitemapStores(_ context.Context, cursor SitemapCursor, limit int) ([]SitemapRecord, int64, error) {
	f.lastSitemapCursor, f.lastSitemapLimit = cursor, limit
	records, total := storeSitemapPage(f.sitemapStores, cursor, limit)
	return records, total, nil
}

// sitemapFixture mirrors one candidate row of the public sitemap scan. Status /
// Approval are only used by the fake to apply the visibility predicate.
type sitemapFixture struct {
	ID        string
	Slug      string
	UpdatedAt time.Time
	Status    string
	Approval  string
}

// productSitemapPage applies the products visibility predicate (status `active`
// AND approval `approved`), mirroring the repository SQL.
func productSitemapPage(rows []sitemapFixture, cursor SitemapCursor, limit int) ([]SitemapRecord, int64) {
	public := make([]SitemapRecord, 0, len(rows))
	for _, fx := range rows {
		if fx.Status == "active" && fx.Approval == "approved" {
			public = append(public, SitemapRecord{ID: fx.ID, Slug: fx.Slug, UpdatedAt: fx.UpdatedAt})
		}
	}
	return sitemapSlice(public, cursor, limit), int64(len(public))
}

// storeSitemapPage applies the stores visibility predicate (status `active`).
func storeSitemapPage(rows []sitemapFixture, cursor SitemapCursor, limit int) ([]SitemapRecord, int64) {
	public := make([]SitemapRecord, 0, len(rows))
	for _, fx := range rows {
		if fx.Status == "active" {
			public = append(public, SitemapRecord{ID: fx.ID, Slug: fx.Slug, UpdatedAt: fx.UpdatedAt})
		}
	}
	return sitemapSlice(public, cursor, limit), int64(len(public))
}

// sitemapSlice orders by (updated_at, id), applies the keyset cursor and clamps
// to limit — the in-memory mirror of the repository's SQL keyset scan.
func sitemapSlice(all []SitemapRecord, cursor SitemapCursor, limit int) []SitemapRecord {
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].UpdatedAt.Equal(all[j].UpdatedAt) {
			return all[i].ID < all[j].ID
		}
		return all[i].UpdatedAt.Before(all[j].UpdatedAt)
	})

	start := len(all)
	for i, rec := range all {
		if afterSitemapCursor(rec, cursor) {
			start = i
			break
		}
	}
	page := all[start:]
	if limit >= 0 && len(page) > limit {
		page = page[:limit]
	}
	return append([]SitemapRecord{}, page...)
}

func afterSitemapCursor(rec SitemapRecord, cursor SitemapCursor) bool {
	if cursor.IsZero() {
		return true
	}
	if rec.UpdatedAt.After(cursor.UpdatedAt) {
		return true
	}
	if rec.UpdatedAt.Equal(cursor.UpdatedAt) {
		return rec.ID > cursor.ID
	}
	return false
}

func sampleProduct() ProductRecord {
	return ProductRecord{
		ID:           "p-1",
		Slug:         "demo-blender",
		Name:         "Demo Blender",
		Price:        39.99,
		Stock:        12,
		Category:     "Home & Kitchen",
		CategorySlug: "home-kitchen",
		Brand:        "FreshMix",
		Rating:       4.5,
		ReviewCount:  3,
		SalesUnits:   7,
		IsNew:        true,
		Attributes:   map[string]string{"shipping_weight": "0.8 kg", "color": "White"},
		CreatedAt:    time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
	}
}
