package storeapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
)

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

func sign(t *testing.T, signer *jwt.Signer, userID, role string) string {
	t.Helper()
	token, err := signer.Sign(userID, role)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return token
}

func do(h http.Handler, method, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func newApplicantRouter(svc *Service, signer *jwt.Signer) http.Handler {
	r := chi.NewRouter()
	r.Route("/api/v1", func(api chi.Router) {
		RegisterApplicant(api, signer, svc, nil)
	})
	return r
}

func newAdminRouter(svc *Service, signer *jwt.Signer, checker auth.RoleChecker) http.Handler {
	r := chi.NewRouter()
	r.Route("/api/v1/admin", func(api chi.Router) {
		api.Use(auth.Authenticate(signer))
		api.Use(auth.RequireRole(checker, "admin"))
		RegisterAdmin(api, svc)
	})
	return r
}

func TestApplicantRoutes_RequireAuth(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	svc := NewService(newFakeRepo(), nil, Config{})
	h := newApplicantRouter(svc, signer)

	if rec := do(h, http.MethodGet, "/api/v1/my-stores", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: got %d, want 401", rec.Code)
	}
	if rec := do(h, http.MethodGet, "/api/v1/my-stores", sign(t, signer, "u-1", "user")); rec.Code != http.StatusOK {
		t.Fatalf("with token: got %d, want 200", rec.Code)
	}
}

func TestAdminApprove_RequiresAdminRole(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	checker := fakeRoleChecker{"admin-1": {"admin"}, "vendor-1": {"vendor"}}

	repo := newFakeRepo()
	repo.app = approvedApp()
	svc := NewService(repo, nil, Config{})
	h := newAdminRouter(svc, signer, checker)

	vendorTok := sign(t, signer, "vendor-1", "vendor")
	if rec := do(h, http.MethodPost, "/api/v1/admin/store-applications/app-1/approve", vendorTok); rec.Code != http.StatusForbidden {
		t.Fatalf("vendor: got %d, want 403", rec.Code)
	}

	adminTok := sign(t, signer, "admin-1", "admin")
	if rec := do(h, http.MethodPost, "/api/v1/admin/store-applications/app-1/approve", adminTok); rec.Code != http.StatusOK {
		t.Fatalf("admin: got %d, want 200", rec.Code)
	}
}

func TestAdminListApplications_PaginationEnvelope(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	checker := fakeRoleChecker{"admin-1": {"admin"}}

	repo := newFakeRepo()
	repo.listAppsTotal = 25
	svc := NewService(repo, nil, Config{})
	h := newAdminRouter(svc, signer, checker)

	rec := do(h, http.MethodGet, "/api/v1/admin/store-applications?page=2&limit=10", sign(t, signer, "admin-1", "admin"))
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !containsAll(body, `"total_pages":3`, `"has_next":true`, `"has_previous":true`) {
		t.Fatalf("unexpected pagination body: %s", body)
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		found := false
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
