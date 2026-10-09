package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
)

func okHandler(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }

// fakeRoleChecker grants roles from a static user -> roles map, standing in for
// the database-backed checker.
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

// TestRequireRole covers the authorization matrix. Roles come from the checker
// (the database), not the token claim, so a user with several roles works too.
func TestRequireRole(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)

	checker := fakeRoleChecker{
		"u-user":   {"user"},
		"u-vendor": {"vendor"},
		"u-admin":  {"admin"},
		"u-both":   {"user", "admin", "vendor"},
	}

	r := chi.NewRouter()
	r.Route("/x", func(r chi.Router) {
		r.Use(Authenticate(signer))
		r.With(RequireRole(checker, "admin")).Get("/admin", okHandler)
		r.With(RequireRole(checker, "vendor", "admin")).Get("/vendor", okHandler)
		r.With(RequireRole(checker, "user", "vendor", "admin")).Get("/user", okHandler)
	})

	token := func(userID, role string) string {
		signed, err := signer.Sign(userID, role)
		if err != nil {
			t.Fatalf("sign: %v", err)
		}
		return signed
	}

	cases := []struct {
		name       string
		path       string
		token      string
		wantStatus int
	}{
		{"user -> user", "/x/user", token("u-user", "user"), http.StatusOK},
		{"user -> admin", "/x/admin", token("u-user", "user"), http.StatusForbidden},
		{"user -> vendor", "/x/vendor", token("u-user", "user"), http.StatusForbidden},
		{"vendor -> vendor", "/x/vendor", token("u-vendor", "vendor"), http.StatusOK},
		{"vendor -> admin", "/x/admin", token("u-vendor", "vendor"), http.StatusForbidden},
		{"admin -> admin", "/x/admin", token("u-admin", "admin"), http.StatusOK},
		{"admin -> user", "/x/user", token("u-admin", "admin"), http.StatusOK},
		// A user holding both roles may operate admin *and* vendor areas.
		{"both -> admin", "/x/admin", token("u-both", "admin"), http.StatusOK},
		{"both -> vendor", "/x/vendor", token("u-both", "admin"), http.StatusOK},
		{"no token", "/x/user", "", http.StatusUnauthorized},
		{"bad token", "/x/user", "not-a-jwt", http.StatusUnauthorized},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, c.path, nil)
			if c.token != "" {
				req.Header.Set("Authorization", "Bearer "+c.token)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != c.wantStatus {
				t.Fatalf("got %d, want %d", rec.Code, c.wantStatus)
			}
		})
	}
}

// TestRequireRole_DeniesWhenCheckerErrors asserts authorization fails closed.
func TestRequireRole_DeniesWhenCheckerErrors(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	token, _ := signer.Sign("u-1", "admin")

	h := Authenticate(signer)(RequireRole(errorChecker{}, "admin")(http.HandlerFunc(okHandler)))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 (fail closed), got %d", rec.Code)
	}
}

type errorChecker struct{}

func (errorChecker) HasAnyRole(context.Context, string, ...string) (bool, error) {
	return false, context.DeadlineExceeded
}

// TestAuthenticate_ExpiredToken asserts an expired token is rejected as 401.
func TestAuthenticate_ExpiredToken(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", -time.Minute)
	token, err := signer.Sign("u-1", "user")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	h := Authenticate(signer)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
