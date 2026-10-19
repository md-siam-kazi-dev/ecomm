package follow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
)

func newRouter(repo FollowRepository, signer *jwt.Signer) http.Handler {
	r := chi.NewRouter()
	mount(r, NewHandler(NewService(repo)), signer)
	return r
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

func signToken(t *testing.T, signer *jwt.Signer, userID string) string {
	t.Helper()
	token, err := signer.Sign(userID, "user")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return token
}

func TestFollowRoutes_AuthRules(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	h := newRouter(&fakeRepo{storeID: "s-1", followerCount: 3}, signer)

	// Guests may read the status + follower count.
	rec := do(h, http.MethodGet, "/stores/demo/follow", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("guest GET: got %d, want 200", rec.Code)
	}
	var body struct {
		Data StatusResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body.Data.Following || body.Data.FollowerCount != 3 {
		t.Fatalf("unexpected guest status: %+v", body.Data)
	}

	// Writes still require auth.
	if rec := do(h, http.MethodPost, "/stores/demo/follow", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("guest POST: got %d, want 401", rec.Code)
	}
	if rec := do(h, http.MethodDelete, "/stores/demo/follow", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("guest DELETE: got %d, want 401", rec.Code)
	}
}

func TestFollow_ThenStatus(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	h := newRouter(&fakeRepo{storeID: "s-1"}, signer)
	token := signToken(t, signer, "u-1")

	rec := do(h, http.MethodPost, "/stores/demo/follow", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("follow: got %d, want 200", rec.Code)
	}
	var body struct {
		Data StatusResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if !body.Data.Following || body.Data.FollowerCount != 1 {
		t.Fatalf("unexpected follow status: %+v", body.Data)
	}

	rec = do(h, http.MethodDelete, "/stores/demo/follow", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("unfollow: got %d, want 200", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body.Data.Following {
		t.Fatalf("expected not following: %+v", body.Data)
	}
}
