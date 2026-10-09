package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRegister_CreatesUserAndHidesPassword(t *testing.T) {
	env := newTestEnv()

	resp := doRequest(env.router(), http.MethodPost, "/auth/register",
		`{"name":"John Doe","email":"John@Example.COM","password":"super-secret-123"}`)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if strings.Contains(body, "super-secret-123") {
		t.Error("plaintext password must never be echoed back")
	}
	if env.users.lastCreate.Email != "john@example.com" {
		t.Errorf("email should be normalized, got %q", env.users.lastCreate.Email)
	}
	if env.users.lastCreate.FirstName != "John" || env.users.lastCreate.LastName != "Doe" {
		t.Errorf("name should split into John/Doe, got %q/%q", env.users.lastCreate.FirstName, env.users.lastCreate.LastName)
	}
	if env.mailer.verifyToken == "" {
		t.Error("a verification email should have been sent")
	}
}

func TestRegister_ValidationFailures(t *testing.T) {
	cases := map[string]string{
		"missing fields": `{"name":"","email":"","password":""}`,
		"invalid email":  `{"name":"John","email":"nope","password":"super-secret-123"}`,
		"weak password":  `{"name":"John","email":"john@example.com","password":"short"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv()
			resp := doRequest(env.router(), http.MethodPost, "/auth/register", body)
			if resp.StatusCode != http.StatusUnprocessableEntity {
				t.Fatalf("expected 422, got %d", resp.StatusCode)
			}
			if env.users.createCalls != 0 {
				t.Error("repository must not be called when validation fails")
			}
		})
	}
}

func TestRegister_RejectsMalformedJSON(t *testing.T) {
	env := newTestEnv()
	resp := doRequest(env.router(), http.MethodPost, "/auth/register", `{not-json`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRegister_DuplicateEmailConflict(t *testing.T) {
	env := newTestEnv()
	env.users.createErr = errEmailTaken

	resp := doRequest(env.router(), http.MethodPost, "/auth/register",
		`{"name":"John Doe","email":"john@example.com","password":"super-secret-123"}`)

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if code := errorCode(t, resp); code != "EMAIL_ALREADY_EXISTS" {
		t.Errorf("expected EMAIL_ALREADY_EXISTS, got %q", code)
	}
}

func TestLogin_SuccessSetsHttpOnlyCookies(t *testing.T) {
	env := newTestEnv()
	env.users.add("john@example.com", "super-secret-123", "user", "active")

	resp := doRequest(env.router(), http.MethodPost, "/auth/login",
		`{"email":"john@example.com","password":"super-secret-123"}`)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, readBody(t, resp))
	}

	access := findCookie(resp, accessCookieName)
	refresh := findCookie(resp, refreshCookieName)
	if access == nil || refresh == nil {
		t.Fatal("expected access and refresh cookies")
	}
	if !access.HttpOnly || !refresh.HttpOnly {
		t.Error("auth cookies must be HttpOnly")
	}
	if refresh.Path != refreshCookiePath {
		t.Errorf("refresh cookie path = %q, want %q", refresh.Path, refreshCookiePath)
	}
	if body := readBody(t, resp); strings.Contains(body, "password") || strings.Contains(body, refresh.Value) {
		t.Error("response must not leak password or refresh token")
	}
}

func TestLogin_FailureModes(t *testing.T) {
	t.Run("wrong password", func(t *testing.T) {
		env := newTestEnv()
		env.users.add("john@example.com", "super-secret-123", "user", "active")
		resp := doRequest(env.router(), http.MethodPost, "/auth/login",
			`{"email":"john@example.com","password":"wrong-password"}`)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("unknown email", func(t *testing.T) {
		env := newTestEnv()
		resp := doRequest(env.router(), http.MethodPost, "/auth/login",
			`{"email":"nobody@example.com","password":"super-secret-123"}`)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
		if code := errorCode(t, resp); code != "INVALID_CREDENTIALS" {
			t.Errorf("expected generic INVALID_CREDENTIALS, got %q", code)
		}
	})

	t.Run("inactive account", func(t *testing.T) {
		env := newTestEnv()
		env.users.add("john@example.com", "super-secret-123", "user", "suspended")
		resp := doRequest(env.router(), http.MethodPost, "/auth/login",
			`{"email":"john@example.com","password":"super-secret-123"}`)
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.StatusCode)
		}
	})
}

func TestRefresh_RotatesAndDetectsReuse(t *testing.T) {
	env := newTestEnv()
	env.users.add("john@example.com", "super-secret-123", "user", "active")
	router := env.router()

	login := doRequest(router, http.MethodPost, "/auth/login",
		`{"email":"john@example.com","password":"super-secret-123"}`)
	firstRefresh := findCookie(login, refreshCookieName)
	if firstRefresh == nil {
		t.Fatal("login did not set a refresh cookie")
	}

	// A successful refresh returns new cookies.
	refresh := doRequest(router, http.MethodPost, "/auth/refresh", "", firstRefresh)
	if refresh.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", refresh.StatusCode, readBody(t, refresh))
	}
	secondRefresh := findCookie(refresh, refreshCookieName)
	if secondRefresh == nil || secondRefresh.Value == firstRefresh.Value {
		t.Fatal("refresh token was not rotated")
	}

	// Reusing the rotated token must fail and revoke the whole family.
	reuse := doRequest(router, http.MethodPost, "/auth/refresh", "", firstRefresh)
	if reuse.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 on reuse, got %d", reuse.StatusCode)
	}
	if active := env.sessions.activeCount("user-1"); active != 0 {
		t.Errorf("expected all sessions revoked after reuse, %d still active", active)
	}
}

func TestLogout_IsIdempotent(t *testing.T) {
	env := newTestEnv()
	env.users.add("john@example.com", "super-secret-123", "user", "active")
	router := env.router()

	login := doRequest(router, http.MethodPost, "/auth/login",
		`{"email":"john@example.com","password":"super-secret-123"}`)
	refresh := findCookie(login, refreshCookieName)

	first := doRequest(router, http.MethodPost, "/auth/logout", "", refresh)
	if first.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", first.StatusCode)
	}
	if active := env.sessions.activeCount("user-1"); active != 0 {
		t.Error("logout should revoke the session")
	}
	// Calling again with the same (now unknown) cookie still succeeds.
	if again := doRequest(router, http.MethodPost, "/auth/logout", "", refresh); again.StatusCode != http.StatusNoContent {
		t.Fatalf("expected idempotent 204, got %d", again.StatusCode)
	}
}

func TestMe_RequiresValidAccessToken(t *testing.T) {
	env := newTestEnv()
	env.users.add("john@example.com", "super-secret-123", "user", "active")
	router := env.router()

	if resp := doRequest(router, http.MethodGet, "/auth/me", ""); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", resp.StatusCode)
	}

	login := doRequest(router, http.MethodPost, "/auth/login",
		`{"email":"john@example.com","password":"super-secret-123"}`)
	access := findCookie(login, accessCookieName)

	resp := doRequest(router, http.MethodGet, "/auth/me", "", access)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if code := errorCode(t, resp); code != "" {
		t.Errorf("expected success envelope, got error %q", code)
	}
}

func TestLogoutAll_RevokesEverySession(t *testing.T) {
	env := newTestEnv()
	env.users.add("john@example.com", "super-secret-123", "user", "active")
	router := env.router()

	// Two separate sessions.
	loginA := doRequest(router, http.MethodPost, "/auth/login", `{"email":"john@example.com","password":"super-secret-123"}`)
	doRequest(router, http.MethodPost, "/auth/login", `{"email":"john@example.com","password":"super-secret-123"}`)
	access := findCookie(loginA, accessCookieName)

	resp := doRequest(router, http.MethodPost, "/auth/logout-all", "", access)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if active := env.sessions.activeCount("user-1"); active != 0 {
		t.Errorf("expected all sessions revoked, %d still active", active)
	}
}

func TestVerifyEmail_ConsumesToken(t *testing.T) {
	env := newTestEnv()
	router := env.router()

	doRequest(router, http.MethodPost, "/auth/register",
		`{"name":"John Doe","email":"john@example.com","password":"super-secret-123"}`)
	if env.mailer.verifyToken == "" {
		t.Fatal("no verification token was issued")
	}

	resp := doRequest(router, http.MethodPost, "/auth/verify-email",
		`{"token":"`+env.mailer.verifyToken+`"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !env.users.verified["user-1"] {
		t.Error("email should be marked verified")
	}

	// Reusing the token fails.
	again := doRequest(router, http.MethodPost, "/auth/verify-email",
		`{"token":"`+env.mailer.verifyToken+`"}`)
	if again.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 on reused token, got %d", again.StatusCode)
	}
}

func TestForgotPassword_GenericAndIssuesToken(t *testing.T) {
	env := newTestEnv()
	env.users.add("john@example.com", "super-secret-123", "user", "active")
	router := env.router()

	resp := doRequest(router, http.MethodPost, "/auth/forgot-password", `{"email":"john@example.com"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if env.mailer.resetToken == "" {
		t.Error("a reset token should have been issued")
	}

	// Unknown email is still a 200 with a generic message.
	unknown := doRequest(router, http.MethodPost, "/auth/forgot-password", `{"email":"nobody@example.com"}`)
	if unknown.StatusCode != http.StatusOK {
		t.Fatalf("expected generic 200, got %d", unknown.StatusCode)
	}
}

func TestResetPassword_UpdatesAndRevokesSessions(t *testing.T) {
	env := newTestEnv()
	env.users.add("john@example.com", "super-secret-123", "user", "active")
	router := env.router()

	login := doRequest(router, http.MethodPost, "/auth/login",
		`{"email":"john@example.com","password":"super-secret-123"}`)
	if login.StatusCode != http.StatusOK {
		t.Fatal("seed login failed")
	}

	doRequest(router, http.MethodPost, "/auth/forgot-password", `{"email":"john@example.com"}`)
	token := env.mailer.resetToken

	resp := doRequest(router, http.MethodPost, "/auth/reset-password",
		`{"token":"`+token+`","new_password":"brand-new-secret"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, readBody(t, resp))
	}
	if active := env.sessions.activeCount("user-1"); active != 0 {
		t.Error("password reset must revoke all sessions")
	}

	// Old password no longer works, new one does.
	oldLogin := doRequest(router, http.MethodPost, "/auth/login",
		`{"email":"john@example.com","password":"super-secret-123"}`)
	if oldLogin.StatusCode != http.StatusUnauthorized {
		t.Errorf("old password should be rejected, got %d", oldLogin.StatusCode)
	}
	newLogin := doRequest(router, http.MethodPost, "/auth/login",
		`{"email":"john@example.com","password":"brand-new-secret"}`)
	if newLogin.StatusCode != http.StatusOK {
		t.Errorf("new password should work, got %d", newLogin.StatusCode)
	}
}

// ---- helpers --------------------------------------------------------------

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(data)
}

func errorCode(t *testing.T, resp *http.Response) string {
	t.Helper()
	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(readBody(t, resp)), &body); err != nil {
		return ""
	}
	return body.Error.Code
}
