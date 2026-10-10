package order

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
	"github.com/md-siam-kazi/e-go-backend/pkg/payment"
)

func newRouter(svc *Service, signer *jwt.Signer) http.Handler {
	r := chi.NewRouter()
	h := NewHandler(svc)

	r.Post("/orders/webhook/{provider}", h.Webhook)
	r.Group(func(r chi.Router) {
		r.Use(auth.Authenticate(signer))
		r.Post("/orders", h.CreateOrder)
		r.Get("/orders", h.ListOrders)
	})
	return r
}

func signToken(t *testing.T, signer *jwt.Signer, userID string) string {
	t.Helper()
	token, err := signer.Sign(userID, "user")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return token
}

func do(h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCreateOrder_RequiresAuth(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	svc := NewService(&fakeRepo{resolved: []VariantResolution{sampleResolution()}}, payment.NewRegistry(), testConfig(""))
	h := newRouter(svc, signer)

	body, _ := json.Marshal(validRequest(MethodCOD))

	if rec := do(h, http.MethodPost, "/orders", "", string(body)); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: got %d, want 401", rec.Code)
	}
	if rec := do(h, http.MethodPost, "/orders", signToken(t, signer, "u-1"), string(body)); rec.Code != http.StatusCreated {
		t.Fatalf("with token: got %d, want 201", rec.Code)
	}
}

func TestListOrders_RequiresAuth(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	svc := NewService(&fakeRepo{}, payment.NewRegistry(), testConfig(""))
	h := newRouter(svc, signer)

	if rec := do(h, http.MethodGet, "/orders", "", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: got %d, want 401", rec.Code)
	}
	if rec := do(h, http.MethodGet, "/orders", signToken(t, signer, "u-1"), ""); rec.Code != http.StatusOK {
		t.Fatalf("with token: got %d, want 200", rec.Code)
	}
}

func TestWebhook_IsPublic(t *testing.T) {
	svc := NewService(&fakeRepo{}, registryWith(&fakeProvider{name: "fake", paid: false}), testConfig("fake"))
	h := newRouter(svc, jwt.NewSigner("secret", "iss", "aud", time.Minute))

	if rec := do(h, http.MethodPost, "/orders/webhook/fake", "", ""); rec.Code != http.StatusOK {
		t.Fatalf("webhook: got %d, want 200", rec.Code)
	}
}

func TestWebhook_NotConfigured(t *testing.T) {
	svc := NewService(&fakeRepo{}, payment.NewRegistry(), testConfig("stripe"))
	h := newRouter(svc, jwt.NewSigner("secret", "iss", "aud", time.Minute))

	if rec := do(h, http.MethodPost, "/orders/webhook/stripe", "", ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503", rec.Code)
	}
}
