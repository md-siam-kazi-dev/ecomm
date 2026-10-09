package catalog

import (
	"context"
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

	related    []ProductRecord
	categories []CategoryRecord
	reviews    []ReviewRecord
	top        []ReviewRecord
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

func (f *fakeRepo) ProductReviews(context.Context, string, int) ([]ReviewRecord, error) {
	return f.reviews, nil
}

func (f *fakeRepo) TopReviews(context.Context, int) ([]ReviewRecord, error) {
	return f.top, nil
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
