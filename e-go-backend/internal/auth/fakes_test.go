package auth

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/config"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
	"github.com/md-siam-kazi/e-go-backend/pkg/password"
)

// ---- fake user repository -------------------------------------------------

type fakeUserRepo struct {
	users       map[string]User
	passwords   map[string]string
	verified    map[string]bool
	createErr   error
	createCalls int
	lastCreate  CreateUserParams
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users:     map[string]User{},
		passwords: map[string]string{},
		verified:  map[string]bool{},
	}
}

func (f *fakeUserRepo) CreateUser(_ context.Context, p CreateUserParams) (CreatedUser, error) {
	f.createCalls++
	if f.createErr != nil {
		return CreatedUser{}, f.createErr
	}
	f.lastCreate = p
	id := "user-" + strconv.Itoa(f.createCalls)
	f.users[id] = User{
		ID: id, Email: p.Email, Name: strings.TrimSpace(p.FirstName + " " + p.LastName),
		Role: "user", Status: "pending",
	}
	f.passwords[id] = p.PasswordHash
	return CreatedUser{ID: id, Email: p.Email, Status: "pending", CreatedAt: time.Now()}, nil
}

func (f *fakeUserRepo) FindUserByEmail(_ context.Context, email string) (User, error) {
	for _, u := range f.users {
		if strings.EqualFold(u.Email, email) {
			return f.hydrate(u), nil
		}
	}
	return User{}, errUserNotFound
}

func (f *fakeUserRepo) FindUserByID(_ context.Context, id string) (User, error) {
	u, ok := f.users[id]
	if !ok {
		return User{}, errUserNotFound
	}
	return f.hydrate(u), nil
}

func (f *fakeUserRepo) UpdatePassword(_ context.Context, userID, hash string) error {
	if _, ok := f.users[userID]; !ok {
		return errUserNotFound
	}
	f.passwords[userID] = hash
	return nil
}

func (f *fakeUserRepo) MarkEmailVerified(_ context.Context, userID string) error {
	f.verified[userID] = true
	if u, ok := f.users[userID]; ok {
		u.Status = "active"
		f.users[userID] = u
	}
	return nil
}

func (f *fakeUserRepo) TouchLastLogin(context.Context, string) error { return nil }

func (f *fakeUserRepo) hydrate(u User) User {
	u.PasswordHash = f.passwords[u.ID]
	u.EmailVerified = f.verified[u.ID]
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	return u
}

// add seeds an account with a hashed password.
func (f *fakeUserRepo) add(email, pw, role, status string) User {
	id := "user-" + strconv.Itoa(len(f.users)+1)
	hash, _ := password.Hash(pw)
	u := User{ID: id, Email: email, Name: "Test User", Role: role, Status: status}
	f.users[id] = u
	f.passwords[id] = hash
	return u
}

// ---- fake session repository ----------------------------------------------

type fakeSessionRepo struct {
	byID   map[string]SessionRecord
	byHash map[string]string
	seq    int
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{byID: map[string]SessionRecord{}, byHash: map[string]string{}}
}

func (f *fakeSessionRepo) Create(_ context.Context, p SessionParams) (SessionRecord, error) {
	f.seq++
	rec := SessionRecord{ID: "sess-" + strconv.Itoa(f.seq), UserID: p.UserID, RefreshHash: p.RefreshHash, ExpiresAt: p.ExpiresAt}
	f.byID[rec.ID] = rec
	f.byHash[p.RefreshHash] = rec.ID
	return rec, nil
}

func (f *fakeSessionRepo) FindByRefreshHash(_ context.Context, hash string) (SessionRecord, error) {
	id, ok := f.byHash[hash]
	if !ok {
		return SessionRecord{}, errSessionNotFound
	}
	return f.byID[id], nil
}

func (f *fakeSessionRepo) Rotate(ctx context.Context, revokeID string, p SessionParams) (SessionRecord, error) {
	if rec, ok := f.byID[revokeID]; ok {
		now := time.Now()
		rec.RevokedAt = &now
		f.byID[revokeID] = rec
	}
	return f.Create(ctx, p)
}

func (f *fakeSessionRepo) RevokeByHash(_ context.Context, hash string) error {
	id, ok := f.byHash[hash]
	if !ok {
		return nil
	}
	rec := f.byID[id]
	now := time.Now()
	rec.RevokedAt = &now
	f.byID[id] = rec
	return nil
}

func (f *fakeSessionRepo) RevokeAllForUser(_ context.Context, userID string) error {
	now := time.Now()
	for id, rec := range f.byID {
		if rec.UserID == userID && rec.RevokedAt == nil {
			rec.RevokedAt = &now
			f.byID[id] = rec
		}
	}
	return nil
}

func (f *fakeSessionRepo) activeCount(userID string) int {
	n := 0
	for _, rec := range f.byID {
		if rec.UserID == userID && rec.RevokedAt == nil {
			n++
		}
	}
	return n
}

// ---- fake token repository ------------------------------------------------

type fakeToken struct {
	userID    string
	purpose   string
	expiresAt time.Time
	usedAt    *time.Time
}

type fakeTokenRepo struct {
	byHash map[string]*fakeToken
}

func newFakeTokenRepo() *fakeTokenRepo { return &fakeTokenRepo{byHash: map[string]*fakeToken{}} }

func (f *fakeTokenRepo) Create(_ context.Context, p TokenParams) error {
	f.byHash[p.TokenHash] = &fakeToken{userID: p.UserID, purpose: p.Purpose, expiresAt: p.ExpiresAt}
	return nil
}

func (f *fakeTokenRepo) Consume(_ context.Context, hash, purpose string) (string, error) {
	tok, ok := f.byHash[hash]
	if !ok || tok.purpose != purpose || tok.usedAt != nil || time.Now().After(tok.expiresAt) {
		return "", errInvalidToken
	}
	now := time.Now()
	tok.usedAt = &now
	return tok.userID, nil
}

func (f *fakeTokenRepo) InvalidateUnused(_ context.Context, userID, purpose string) error {
	now := time.Now()
	for _, tok := range f.byHash {
		if tok.userID == userID && tok.purpose == purpose && tok.usedAt == nil {
			tok.usedAt = &now
		}
	}
	return nil
}

// ---- fake mailer ----------------------------------------------------------

type captureMailer struct {
	verifyTo, verifyToken string
	resetTo, resetToken   string
}

func (m *captureMailer) SendVerificationEmail(_ context.Context, to, token string) error {
	m.verifyTo, m.verifyToken = to, token
	return nil
}

func (m *captureMailer) SendPasswordResetEmail(_ context.Context, to, token string) error {
	m.resetTo, m.resetToken = to, token
	return nil
}

// ---- test environment -----------------------------------------------------

type testEnv struct {
	users     *fakeUserRepo
	sessions  *fakeSessionRepo
	tokens    *fakeTokenRepo
	mailer    *captureMailer
	signer    *jwt.Signer
	svc       *Service
	handler   *Handler
	authCfg   config.AuthConfig
	accessTTL time.Duration
}

func newTestEnv() *testEnv {
	users := newFakeUserRepo()
	sessions := newFakeSessionRepo()
	tokens := newFakeTokenRepo()
	mailer := &captureMailer{}
	signer := jwt.NewSigner("test-secret", "test-issuer", "test-audience", 15*time.Minute)

	accessTTL := 15 * time.Minute
	refreshTTL := 7 * 24 * time.Hour

	svc := NewService(ServiceDeps{
		Users: users, Sessions: sessions, Tokens: tokens, Signer: signer,
		Mailer: mailer, Log: slog.Default(), AccessTTL: accessTTL, RefreshTTL: refreshTTL,
	})
	cfg := config.AuthConfig{}
	h := NewHandler(svc, cfg, accessTTL, refreshTTL)

	return &testEnv{users: users, sessions: sessions, tokens: tokens, mailer: mailer, signer: signer, svc: svc, handler: h, authCfg: cfg, accessTTL: accessTTL}
}

func (e *testEnv) router() http.Handler {
	r := chi.NewRouter()
	mount(r, e.handler, e.signer, nil)
	return r
}

// ---- request helpers ------------------------------------------------------

func doRequest(h http.Handler, method, path, body string, cookies ...*http.Cookie) *http.Response {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Result()
}

func findCookie(resp *http.Response, name string) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}
