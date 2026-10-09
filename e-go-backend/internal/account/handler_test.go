package account

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
)

func newRouter(repo Store, signer *jwt.Signer) http.Handler {
	return newRouterWithUploader(repo, &fakeUploader{}, signer)
}

func newRouterWithUploader(repo Store, uploader ImageUploader, signer *jwt.Signer) http.Handler {
	r := chi.NewRouter()
	r.Route("/account", func(r chi.Router) {
		r.Use(auth.Authenticate(signer))
		mount(r, NewHandler(NewService(repo, uploader)))
	})
	return r
}

func request(h http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	var req *http.Request
	if reader != nil {
		req = httptest.NewRequest(method, path, reader)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func userToken(t *testing.T, signer *jwt.Signer) string {
	t.Helper()
	token, err := signer.Sign("u-1", "user")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return token
}

func TestAddresses_RequireAuth(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	h := newRouter(&fakeRepo{}, signer)

	if rec := request(h, http.MethodGet, "/account/addresses", "", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", rec.Code)
	}
}

func TestAddresses_CRUD(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	repo := &fakeRepo{}
	h := newRouter(repo, signer)
	token := userToken(t, signer)

	// Create
	create := request(h, http.MethodPost, "/account/addresses", `{
		"full_name":"Siam Kazi","phone":"+8801700000000",
		"division":"Dhaka","district":"Gazipur","upazila":"Kaliakair","village_town":"Mouchak"
	}`, token)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", create.Code, create.Body.String())
	}
	var created struct {
		Data AddressResponse `json:"data"`
	}
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if created.Data.District != "Gazipur" || !created.Data.IsDefault {
		t.Fatalf("unexpected created address: %+v", created.Data)
	}

	// List
	list := request(h, http.MethodGet, "/account/addresses", "", token)
	if list.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", list.Code)
	}

	// Update
	update := request(h, http.MethodPut, "/account/addresses/a-1",
		`{"full_name":"Siam Kazi","phone":"+8801700000000","division":"Dhaka","district":"Dhaka","upazila":"Savar","village_town":"Ashulia"}`, token)
	if update.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", update.Code)
	}
	if repo.lastID != "a-1" || repo.lastParam.District != "Dhaka" {
		t.Errorf("update not forwarded: id=%q param=%+v", repo.lastID, repo.lastParam)
	}

	// Set default
	def := request(h, http.MethodPost, "/account/addresses/a-1/default", "", token)
	if def.Code != http.StatusOK || !repo.defaulted {
		t.Fatalf("expected 200 and default set, got %d defaulted=%v", def.Code, repo.defaulted)
	}

	// Delete
	del := request(h, http.MethodDelete, "/account/addresses/a-1", "", token)
	if del.Code != http.StatusNoContent || !repo.deleted {
		t.Fatalf("expected 204, got %d deleted=%v", del.Code, repo.deleted)
	}
}

func TestProfile_UpdateAndRead(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	repo := &fakeRepo{}
	h := newRouter(repo, signer)
	token := userToken(t, signer)

	update := request(h, http.MethodPut, "/account/profile",
		`{"full_name":"Siam Kazi","phone":"+8801700000000"}`, token)
	if update.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", update.Code, update.Body.String())
	}
	if repo.profileFullName != "Siam Kazi" || repo.profilePhone != "+8801700000000" {
		t.Errorf("profile not persisted: %q %q", repo.profileFullName, repo.profilePhone)
	}

	read := request(h, http.MethodGet, "/account/profile", "", token)
	if read.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", read.Code)
	}
}

func TestProfile_RejectsInvalid(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	h := newRouter(&fakeRepo{}, signer)

	rec := request(h, http.MethodPut, "/account/profile", `{"full_name":"X","phone":""}`, userToken(t, signer))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestUploadAvatar_HostsAndStores(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	repo := &fakeRepo{}
	uploader := &fakeUploader{url: "https://i.ibb.co/abc/pic.png"}
	h := newRouterWithUploader(repo, uploader, signer)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("image", "pic.png")
	// A tiny PNG header so content-type detection accepts it.
	_, _ = part.Write([]byte("\x89PNG\r\n\x1a\n..."))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/account/avatar", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+userToken(t, signer))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if repo.avatarURL != "https://i.ibb.co/abc/pic.png" {
		t.Errorf("avatar url not stored: %q", repo.avatarURL)
	}
}

func TestUploadAvatar_RejectsMissingFile(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	h := newRouter(&fakeRepo{}, signer)

	req := httptest.NewRequest(http.MethodPost, "/account/avatar", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	req.Header.Set("Authorization", "Bearer "+userToken(t, signer))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAddresses_RejectInvalidPayload(t *testing.T) {
	signer := jwt.NewSigner("secret", "iss", "aud", time.Minute)
	h := newRouter(&fakeRepo{}, signer)

	rec := request(h, http.MethodPost, "/account/addresses",
		`{"full_name":"Siam Kazi","phone":"+8801700000000","division":"","district":"","upazila":"","village_town":""}`,
		userToken(t, signer))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}
