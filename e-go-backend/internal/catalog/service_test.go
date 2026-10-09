package catalog

import (
	"context"
	"errors"
	"testing"
)

func TestListProducts_NormalizesQuery(t *testing.T) {
	repo := &fakeRepo{}
	_, pagination, err := NewService(repo).ListProducts(context.Background(), ListQuery{
		Page: 0, Limit: 999, Sort: "bogus", Search: "  blender  ",
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pagination.Page != 1 || pagination.Limit != maxLimit {
		t.Errorf("pagination not normalized: %+v", pagination)
	}
	if repo.lastQuery.Sort != "featured" {
		t.Errorf("invalid sort should fall back to featured, got %q", repo.lastQuery.Sort)
	}
	if repo.lastQuery.Search != "blender" {
		t.Errorf("search not trimmed: %q", repo.lastQuery.Search)
	}
}

func TestListProducts_DefaultLimit(t *testing.T) {
	repo := &fakeRepo{}
	if _, pagination, _ := NewService(repo).ListProducts(context.Background(), ListQuery{}, nil); pagination.Limit != defaultLimit {
		t.Errorf("limit = %d, want %d", pagination.Limit, defaultLimit)
	}
}

func TestGetProduct_DerivesFlagsAndSpecs(t *testing.T) {
	repo := &fakeRepo{product: sampleProduct()}
	dto, err := NewService(repo).GetProduct(context.Background(), "demo-blender")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dto.IsFeatured {
		t.Error("rating >= 4.0 should be featured")
	}
	if !dto.IsBestSeller {
		t.Error("sales > 0 should be a best seller")
	}
	if len(dto.Specifications) != 2 {
		t.Fatalf("expected 2 specs, got %+v", dto.Specifications)
	}
	// Sorted by raw key: color, shipping_weight.
	if dto.Specifications[0].Label != "Color" || dto.Specifications[0].Value != "White" {
		t.Errorf("unexpected first spec: %+v", dto.Specifications[0])
	}
	if dto.Specifications[1].Label != "Shipping weight" {
		t.Errorf("key not humanized: %+v", dto.Specifications[1])
	}
}

func TestGetProduct_NotFeaturedBelowThreshold(t *testing.T) {
	rec := sampleProduct()
	rec.Rating = 3.2
	rec.SalesUnits = 0
	repo := &fakeRepo{product: rec}

	dto, err := NewService(repo).GetProduct(context.Background(), "demo-blender")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto.IsFeatured || dto.IsBestSeller {
		t.Errorf("expected not featured/best-seller, got %+v", dto)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	repo := &fakeRepo{productErr: errProductNotFound}
	_, err := NewService(repo).GetProduct(context.Background(), "missing")
	if !errors.Is(err, errProductNotFound) {
		t.Fatalf("expected errProductNotFound, got %v", err)
	}
}

func TestCategories_MapsAndImageAlt(t *testing.T) {
	repo := &fakeRepo{categories: []CategoryRecord{
		{ID: "c-1", Slug: "electronics", Name: "Electronics & Gadgets", ProductCount: 4},
	}}
	items, err := NewService(repo).Categories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ImageAlt != "Electronics & Gadgets category" {
		t.Fatalf("unexpected categories: %+v", items)
	}
}

func TestListProducts_PassesIDs(t *testing.T) {
	repo := &fakeRepo{}
	_, _, err := NewService(repo).ListProducts(context.Background(), ListQuery{}, []string{"a", "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.lastIDs) != 2 {
		t.Errorf("ids not forwarded: %+v", repo.lastIDs)
	}
}
