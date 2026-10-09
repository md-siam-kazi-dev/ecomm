package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func newRouter(repo CatalogRepository) http.Handler {
	r := chi.NewRouter()
	mount(r, NewHandler(NewService(repo)))
	return r
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestListProducts_PaginationEnvelope(t *testing.T) {
	repo := &fakeRepo{products: []ProductRecord{sampleProduct()}, total: 25}
	rec := get(newRouter(repo), "/products?page=2&limit=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}

	var body struct {
		Data       []ProductDTO `json:"data"`
		Pagination struct {
			Page       int  `json:"page"`
			Limit      int  `json:"limit"`
			Total      int  `json:"total"`
			TotalPages int  `json:"total_pages"`
			HasNext    bool `json:"has_next"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body.Pagination.TotalPages != 3 || !body.Pagination.HasNext || body.Pagination.Page != 2 {
		t.Fatalf("unexpected pagination: %+v", body.Pagination)
	}
	if len(body.Data) != 1 || body.Data[0].Slug != "demo-blender" {
		t.Fatalf("unexpected data: %+v", body.Data)
	}
}

func TestGetProduct_RouteNotFound(t *testing.T) {
	repo := &fakeRepo{productErr: errProductNotFound}
	rec := get(newRouter(repo), "/products/missing")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404", rec.Code)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != "PRODUCT_NOT_FOUND" {
		t.Errorf("expected PRODUCT_NOT_FOUND, got %q", body.Error.Code)
	}
}

func TestListCategories(t *testing.T) {
	repo := &fakeRepo{categories: []CategoryRecord{{ID: "c-1", Slug: "electronics", Name: "Electronics"}}}
	rec := get(newRouter(repo), "/categories")
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
}
