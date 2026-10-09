package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/internal/config"
)

// Verifies the auth module is mounted under /api/v1 through the real router and
// middleware stack. A malformed body is rejected by the handler before any
// database access, so the server can be built without live Postgres/Redis.
func TestAuthRoutes_AreMounted(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("JWT_ACCESS_SECRET", "test-access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load: %v", err)
	}

	srv := New(cfg, slog.Default(), nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
