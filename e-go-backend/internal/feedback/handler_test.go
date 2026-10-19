package feedback

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
)

func newRouter(repo FeedbackRepository, signer *jwt.Signer) http.Handler {
	r := chi.NewRouter()
	mount(r, NewHandler(NewService(repo, &fakeUploader{})), signer, nil)
	return r
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

func signToken(t *testing.T, signer *jwt.Signer, userID string) string {
	t.Helper()
	token, err := signer.Sign(userID, "user")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return token
}

func TestCreateReview_RequiresAuthAndCreates(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	repo := &fakeRepo{elig: OrderItemEligibility{ProductID: "p-1", UserID: "u-1", OrderStatus: "delivered"}}
	h := newRouter(repo, signer)

	body := `{"order_item_id":"oi-1","rating":5,"title":"Great","comment":"Loved it"}`
	if rec := do(h, http.MethodPost, "/reviews", "", body); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: got %d, want 401", rec.Code)
	}

	rec := do(h, http.MethodPost, "/reviews", signToken(t, signer, "u-1"), body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("with token: got %d, want 201 (%s)", rec.Code, rec.Body.String())
	}

	var payload struct {
		Success bool           `json:"success"`
		Data    ReviewResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if !payload.Success || payload.Data.Rating != 5 || payload.Data.ProductID != "p-1" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestCreateReview_IneligibleReturns422(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	repo := &fakeRepo{elig: OrderItemEligibility{ProductID: "p-1", UserID: "u-1", OrderStatus: "pending"}}
	h := newRouter(repo, signer)

	rec := do(h, http.MethodPost, "/reviews", signToken(t, signer, "u-1"),
		`{"order_item_id":"oi-1","rating":5,"comment":"Loved it"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got %d, want 422", rec.Code)
	}
}

func TestUpdateReview_RequiresAuth(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	repo := &fakeRepo{updated: ReviewRecord{ID: "r-1", Rating: 4, Comment: "updated"}}
	h := newRouter(repo, signer)

	body := `{"rating":4,"comment":"updated"}`
	if rec := do(h, http.MethodPut, "/reviews/r-1", "", body); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: got %d, want 401", rec.Code)
	}
	if rec := do(h, http.MethodPut, "/reviews/r-1", signToken(t, signer, "u-1"), body); rec.Code != http.StatusOK {
		t.Fatalf("with token: got %d, want 200", rec.Code)
	}
}

func TestCreateQuestion_RequiresAuthAndCreates(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	repo := &fakeRepo{productID: "p-1"}
	h := newRouter(repo, signer)

	body := `{"question":"Is it dishwasher safe?"}`
	if rec := do(h, http.MethodPost, "/products/demo-blender/questions", "", body); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: got %d, want 401", rec.Code)
	}
	if rec := do(h, http.MethodPost, "/products/demo-blender/questions", signToken(t, signer, "u-2"), body); rec.Code != http.StatusCreated {
		t.Fatalf("with token: got %d, want 201 (%s)", rec.Code, rec.Body.String())
	}
}
