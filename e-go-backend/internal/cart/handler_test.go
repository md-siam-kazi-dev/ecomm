package cart

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/config"
	pkgjwt "github.com/md-siam-kazi/e-go-backend/pkg/jwt"
)

func newRouter(repo CartRepository) (http.Handler, *pkgjwt.Signer) {
	signer := pkgjwt.NewSigner("secret", "iss", "aud", time.Minute)
	r := chi.NewRouter()
	mount(r, NewHandler(NewService(repo, "usd"), config.AuthConfig{}), signer, nil)
	return r, signer
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

func findCookie(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestGet_Guest_IssuesSessionCookie(t *testing.T) {
	h, _ := newRouter(&fakeRepo{})

	rec := do(h, http.MethodGet, "/cart", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	if cookie := findCookie(rec, CookieName); cookie == nil || cookie.Value == "" {
		t.Fatal("expected a non-empty cart_session cookie for a guest")
	}
}

func TestAdd_AuthenticatedCallerUsesUserCart(t *testing.T) {
	repo := &fakeRepo{
		userCart:    Cart{ID: "u-cart", UserID: "u-1", Currency: "usd"},
		hasUserCart: true,
		variant:     VariantResolution{ProductID: "p-1", VariantID: "v-1", UnitPrice: 9.99, Available: 4},
	}
	h, signer := newRouter(repo)
	token, err := signer.Sign("u-1", "customer")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	rec := do(h, http.MethodPost, "/cart/items", token, `{"product_id":"p-1","quantity":2}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	if repo.lastUpsert.cartID != "u-cart" {
		t.Fatalf("upsert cart = %q, want u-cart", repo.lastUpsert.cartID)
	}

	var payload struct {
		Success bool         `json:"success"`
		Data    CartResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if !payload.Success {
		t.Fatal("expected success envelope")
	}
}

func TestSetQuantity_RoutesItemID(t *testing.T) {
	repo := &fakeRepo{
		userCart:    Cart{ID: "u-cart", UserID: "u-1", Currency: "usd"},
		hasUserCart: true,
		items:       []ItemRecord{{ID: "i-9", Available: 5}},
	}
	h, signer := newRouter(repo)
	token, _ := signer.Sign("u-1", "customer")

	rec := do(h, http.MethodPut, "/cart/items/i-9", token, `{"quantity":3}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	if repo.lastSetQty.itemID != "i-9" || repo.lastSetQty.quantity != 3 {
		t.Fatalf("unexpected set quantity: %+v", repo.lastSetQty)
	}
}

func TestRemove_RoutesItemID(t *testing.T) {
	repo := &fakeRepo{
		userCart:    Cart{ID: "u-cart", UserID: "u-1", Currency: "usd"},
		hasUserCart: true,
	}
	h, signer := newRouter(repo)
	token, _ := signer.Sign("u-1", "customer")

	rec := do(h, http.MethodDelete, "/cart/items/i-3", token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	if repo.removedItem != "i-3" {
		t.Fatalf("removed item = %q, want i-3", repo.removedItem)
	}
}
