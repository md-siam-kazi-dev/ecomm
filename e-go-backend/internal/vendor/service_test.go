package vendor

import (
	"context"
	"errors"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

func validProductRequest() ProductRequest {
	return ProductRequest{
		Name:        "Portable Blender",
		Description: "A compact blender for smoothies.",
		Category:    "electronics",
		Price:       39.99,
		Stock:       10,
		Images:      []string{"https://i.ibb.co/x/blender.png"},
		Attributes:  map[string]string{"color": "Black"},
	}
}

func TestCreateProduct_Validation(t *testing.T) {
	cases := map[string]func(*ProductRequest){
		"missing name":        func(r *ProductRequest) { r.Name = "" },
		"missing description": func(r *ProductRequest) { r.Description = "" },
		"missing category":    func(r *ProductRequest) { r.Category = "" },
		"zero price":          func(r *ProductRequest) { r.Price = 0 },
		"negative price":      func(r *ProductRequest) { r.Price = -1 },
		"sale not below":      func(r *ProductRequest) { p := r.Price; r.SalePrice = &p },
		"sale above":          func(r *ProductRequest) { p := r.Price + 5; r.SalePrice = &p },
		"negative stock":      func(r *ProductRequest) { r.Stock = -1 },
		"no images":           func(r *ProductRequest) { r.Images = nil },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			req := validProductRequest()
			mutate(&req)

			_, err := NewService(newFakeRepo(), &fakeUploader{}).CreateProduct(context.Background(), "u-1", "", req)
			appErr, ok := apperr.As(err)
			if !ok || appErr.Code != apperr.ErrValidation.Code {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestCreateProduct_ValidPassesParams(t *testing.T) {
	repo := newFakeRepo()
	req := validProductRequest()
	req.Submit = true
	req.Attributes = map[string]string{"color": "  Black  ", "warranty": ""}

	_, err := NewService(repo, &fakeUploader{}).CreateProduct(context.Background(), "u-1", "", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.lastProductParams.Submit {
		t.Error("submit flag should be forwarded")
	}
	if repo.lastProductParams.Attributes["color"] != "Black" {
		t.Errorf("attributes not trimmed: %+v", repo.lastProductParams.Attributes)
	}
	if _, exists := repo.lastProductParams.Attributes["warranty"]; exists {
		t.Error("empty attribute values should be dropped")
	}
}

func TestUploadImage_StoresHostedURL(t *testing.T) {
	uploader := &fakeUploader{url: "https://i.ibb.co/x/pic.png"}
	url, err := NewService(newFakeRepo(), uploader).UploadImage(context.Background(), "pic.png", []byte("bytes"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://i.ibb.co/x/pic.png" || !uploader.called {
		t.Fatalf("unexpected url %q called=%v", url, uploader.called)
	}
}

func TestListProducts_NormalizesPagination(t *testing.T) {
	cases := []struct {
		name      string
		in        ListQuery
		wantPage  int
		wantLimit int
	}{
		{"defaults", ListQuery{}, 1, defaultLimit},
		{"invalid limit", ListQuery{Limit: 7}, 1, defaultLimit},
		{"allowed limit", ListQuery{Limit: 50}, 1, 50},
		{"negative page", ListQuery{Page: -4}, 1, defaultLimit},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := newFakeRepo()
			_, pagination, err := NewService(repo, &fakeUploader{}).ListProducts(context.Background(), "u-1", "", c.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pagination.Page != c.wantPage || pagination.Limit != c.wantLimit {
				t.Errorf("page/limit = %d/%d, want %d/%d",
					pagination.Page, pagination.Limit, c.wantPage, c.wantLimit)
			}
			if repo.lastProductQuery.Limit != c.wantLimit {
				t.Errorf("repo limit = %d, want %d", repo.lastProductQuery.Limit, c.wantLimit)
			}
		})
	}
}

func TestListOrders_ValidatesStatusFilter(t *testing.T) {
	repo := newFakeRepo()
	if _, _, err := NewService(repo, &fakeUploader{}).ListOrders(context.Background(), "u-1", "", ListQuery{Status: "NONSENSE"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastOrderQuery.Status != "" {
		t.Errorf("invalid status should be dropped, got %q", repo.lastOrderQuery.Status)
	}

	repo2 := newFakeRepo()
	if _, _, err := NewService(repo2, &fakeUploader{}).ListOrders(context.Background(), "u-1", "", ListQuery{Status: "Shipped"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo2.lastOrderQuery.Status != "shipped" {
		t.Errorf("status should be normalized to lowercase, got %q", repo2.lastOrderQuery.Status)
	}
}

func TestProfile_RequiresVendorProfile(t *testing.T) {
	repo := newFakeRepo()
	repo.scopeErr = errNoActiveStore

	_, err := NewService(repo, &fakeUploader{}).Profile(context.Background(), "u-1", "")
	if !errors.Is(err, errNoActiveStore) {
		t.Fatalf("expected errNoActiveStore, got %v", err)
	}
}

func TestDashboard_MapsSummary(t *testing.T) {
	repo := newFakeRepo()
	repo.summary = Summary{Revenue: 1234.5, OrderCount: 9, ProductCount: 4, LowStockCount: 2, PendingFulfilment: 3}

	got, err := NewService(repo, &fakeUploader{}).Dashboard(context.Background(), "u-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Revenue != 1234.5 || got.PendingFulfilment != 3 || got.LowStockCount != 2 {
		t.Fatalf("unexpected summary: %+v", got)
	}
}

func TestProfile_MapsVendor(t *testing.T) {
	got, err := NewService(newFakeRepo(), &fakeUploader{}).Profile(context.Background(), "u-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.BusinessName != "Vendor Demo Co" || got.Status != "active" {
		t.Fatalf("unexpected profile: %+v", got)
	}
}

func TestListInventory_NormalizesStockFilter(t *testing.T) {
	repo := newFakeRepo()
	_, pagination, err := NewService(repo, &fakeUploader{}).ListInventory(context.Background(), "u-1", "", ListQuery{Status: "NONSENSE", Limit: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastInventoryQuery.Status != "" {
		t.Errorf("invalid stock filter should be dropped, got %q", repo.lastInventoryQuery.Status)
	}
	if pagination.Limit != defaultLimit {
		t.Errorf("invalid limit should fall back to default, got %d", pagination.Limit)
	}

	repo2 := newFakeRepo()
	if _, _, err := NewService(repo2, &fakeUploader{}).ListInventory(context.Background(), "u-1", "", ListQuery{Status: "LOW"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo2.lastInventoryQuery.Status != "low" {
		t.Errorf("stock filter should be normalized to lowercase, got %q", repo2.lastInventoryQuery.Status)
	}
}

func TestAdjustStock_RejectsNegative(t *testing.T) {
	_, err := NewService(newFakeRepo(), &fakeUploader{}).AdjustStock(context.Background(), "u-1", "", "v-1", InventoryAdjustRequest{Quantity: -1})
	appErr, ok := apperr.As(err)
	if !ok || appErr.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestAdjustStock_MapsItem(t *testing.T) {
	repo := newFakeRepo()
	repo.adjusted = InventoryRecord{
		VariantID: "v-1", ProductID: "p-1", Name: "Blender", SKU: "BLENDER-1",
		Category: "home-kitchen", Quantity: 20, Reserved: 2,
	}

	item, err := NewService(repo, &fakeUploader{}).AdjustStock(context.Background(), "u-1", "", "v-1", InventoryAdjustRequest{Quantity: 20, Note: "  restock  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Available != 18 {
		t.Errorf("available = %d, want 18", item.Available)
	}
	if repo.lastAdjustVariant != "v-1" || repo.lastAdjustQuantity != 20 {
		t.Errorf("adjust not forwarded: variant=%q quantity=%d", repo.lastAdjustVariant, repo.lastAdjustQuantity)
	}
}
