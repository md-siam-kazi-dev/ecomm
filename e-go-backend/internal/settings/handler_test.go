package settings

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	pkgjwt "github.com/md-siam-kazi/e-go-backend/pkg/jwt"
)

func newAdminRouter(repo SettingsRepository, signer *pkgjwt.Signer) http.Handler {
	r := chi.NewRouter()
	r.Use(auth.Authenticate(signer))
	mountAdmin(r, NewHandler(NewService(repo)))
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

func token(t *testing.T, signer *pkgjwt.Signer) string {
	t.Helper()
	signed, err := signer.Sign("u-1", "admin")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}

func TestPublicGet_ReturnsSettings(t *testing.T) {
	repo := &fakeRepo{get: Settings{
		AnnouncementEnabled: true,
		AnnouncementMessage: "Hello",
		UpdatedAt:           time.Now(),
	}}
	r := chi.NewRouter()
	mountPublic(r, NewHandler(NewService(repo)))

	rec := do(r, http.MethodGet, "/storefront/settings", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}

	var payload struct {
		Success bool `json:"success"`
		Data    Item `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if !payload.Success || payload.Data.AnnouncementMessage != "Hello" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestAdminUpdate_RequiresAuth(t *testing.T) {
	signer := pkgjwt.NewSigner("secret", "iss", "aud", time.Minute)
	h := newAdminRouter(&fakeRepo{}, signer)

	body := `{"announcement_enabled":true,"announcement_message":"Hi"}`
	if rec := do(h, http.MethodPut, "/settings", "", body); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: got %d, want 401", rec.Code)
	}
}

func TestAdminUpdate_Persists(t *testing.T) {
	signer := pkgjwt.NewSigner("secret", "iss", "aud", time.Minute)
	repo := &fakeRepo{updated: Settings{AnnouncementEnabled: true, AnnouncementMessage: "Hi"}}
	h := newAdminRouter(repo, signer)

	body := `{"announcement_enabled":true,"announcement_message":"Hi"}`
	rec := do(h, http.MethodPut, "/settings", token(t, signer), body)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	if repo.lastActor != "u-1" || repo.lastMsg != "Hi" {
		t.Fatalf("unexpected repo args: actor=%q msg=%q", repo.lastActor, repo.lastMsg)
	}
}

func TestAdminUpdate_RejectsBadJSON(t *testing.T) {
	signer := pkgjwt.NewSigner("secret", "iss", "aud", time.Minute)
	h := newAdminRouter(&fakeRepo{}, signer)

	rec := do(h, http.MethodPut, "/settings", token(t, signer), "{not json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", rec.Code)
	}
}
