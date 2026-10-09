package vendor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
)

// fakeRoleChecker grants roles from a static user -> roles map (stands in for
// the database-backed checker).
type fakeRoleChecker map[string][]string

func (f fakeRoleChecker) HasAnyRole(_ context.Context, userID string, roles ...string) (bool, error) {
	for _, want := range roles {
		for _, have := range f[userID] {
			if want == have {
				return true, nil
			}
		}
	}
	return false, nil
}

// newRouter builds the real auth + role middleware around the vendor handlers
// so authorization is exercised end to end (with a fake repository).
func newRouter(repo VendorRepository, signer *jwt.Signer, checker auth.RoleChecker) http.Handler {
	r := chi.NewRouter()
	r.Route("/vendor", func(r chi.Router) {
		r.Use(auth.Authenticate(signer))
		r.Use(auth.RequireRole(checker, "vendor"))
		mount(r, NewHandler(NewService(repo, &fakeUploader{})))
	})
	return r
}

func doGet(h http.Handler, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func sign(t *testing.T, signer *jwt.Signer, userID, role string) string {
	t.Helper()
	token, err := signer.Sign(userID, role)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return token
}

func TestVendorRoutes_RequireVendorRole(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	checker := fakeRoleChecker{
		"vendor-user": {"vendor"},
		"admin-user":  {"admin"},
		"user-user":   {"user"},
	}
	h := newRouter(newFakeRepo(), signer, checker)

	cases := []struct {
		name   string
		token  string
		status int
	}{
		{"no token", "", http.StatusUnauthorized},
		{"vendor", sign(t, signer, "vendor-user", "vendor"), http.StatusOK},
		{"admin", sign(t, signer, "admin-user", "admin"), http.StatusForbidden},
		{"user", sign(t, signer, "user-user", "user"), http.StatusForbidden},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := doGet(h, "/vendor/dashboard", c.token)
			if rec.Code != c.status {
				t.Fatalf("got %d, want %d", rec.Code, c.status)
			}
		})
	}
}

func TestVendorDashboard_ReturnsSummary(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	checker := fakeRoleChecker{"u-1": {"vendor"}}

	repo := newFakeRepo()
	repo.summary = Summary{Revenue: 500, OrderCount: 3, ProductCount: 2, LowStockCount: 1, PendingFulfilment: 2}

	rec := doGet(newRouter(repo, signer, checker), "/vendor/dashboard", sign(t, signer, "u-1", "vendor"))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Success bool            `json:"success"`
		Data    SummaryResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body.Data.Revenue != 500 || body.Data.PendingFulfilment != 2 {
		t.Fatalf("unexpected summary: %+v", body.Data)
	}
}

func TestVendorProducts_PaginationMetadata(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	checker := fakeRoleChecker{"u-1": {"vendor"}}

	repo := newFakeRepo()
	repo.productsTotal = 25

	rec := doGet(newRouter(repo, signer, checker), "/vendor/products?page=2&limit=10", sign(t, signer, "u-1", "vendor"))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Data       []ProductItem `json:"data"`
		Pagination struct {
			Page        int  `json:"page"`
			Limit       int  `json:"limit"`
			Total       int  `json:"total"`
			TotalPages  int  `json:"total_pages"`
			HasNext     bool `json:"has_next"`
			HasPrevious bool `json:"has_previous"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body.Pagination.Page != 2 || body.Pagination.Limit != 10 || body.Pagination.TotalPages != 3 {
		t.Fatalf("unexpected pagination: %+v", body.Pagination)
	}
	if !body.Pagination.HasNext || !body.Pagination.HasPrevious {
		t.Errorf("has_next/has_previous = %v/%v, want true/true",
			body.Pagination.HasNext, body.Pagination.HasPrevious)
	}
}

func TestVendorProfile_NoVendorRow(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	checker := fakeRoleChecker{"u-1": {"vendor"}}

	repo := newFakeRepo()
	repo.scopeErr = errNoActiveStore

	rec := doGet(newRouter(repo, signer, checker), "/vendor/profile", sign(t, signer, "u-1", "vendor"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != "NO_ACTIVE_STORE" {
		t.Errorf("expected NO_ACTIVE_STORE, got %q", body.Error.Code)
	}
}
