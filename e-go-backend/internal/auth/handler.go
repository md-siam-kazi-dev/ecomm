package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/md-siam-kazi/e-go-backend/internal/config"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/middleware"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

var errInvalidBody = apperr.New(http.StatusBadRequest, "BAD_REQUEST", "Request body must be valid JSON")

// Handler translates auth HTTP requests into service calls. It holds no
// business logic (PRD §4).
type Handler struct {
	svc        *Service
	authCfg    config.AuthConfig
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewHandler(svc *Service, authCfg config.AuthConfig, accessTTL, refreshTTL time.Duration) *Handler {
	return &Handler{svc: svc, authCfg: authCfg, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// Register handles POST /api/v1/auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	ack, err := h.svc.Register(r.Context(), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Created(w, ack, "Account created successfully")
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	result, err := h.svc.Login(r.Context(), req, clientMeta(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}

	h.writeSessionCookies(w, result)
	response.OK(w, userEnvelope{User: toUserResponse(result.User)}, "Signed in")
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.Refresh(r.Context(), refreshTokenFromCookie(r), clientMeta(r))
	if err != nil {
		clearAuthCookies(w, h.authCfg)
		response.Error(w, r, err)
		return
	}

	h.writeSessionCookies(w, result)
	response.OK(w, userEnvelope{User: toUserResponse(result.User)}, "Session refreshed")
}

// Logout handles POST /api/v1/auth/logout. Safe to call repeatedly.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Logout(r.Context(), refreshTokenFromCookie(r)); err != nil {
		response.Error(w, r, err)
		return
	}
	clearAuthCookies(w, h.authCfg)
	response.NoContent(w)
}

// LogoutAll handles POST /api/v1/auth/logout-all (authenticated).
func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	if err := h.svc.LogoutAll(r.Context(), identity.UserID); err != nil {
		response.Error(w, r, err)
		return
	}
	clearAuthCookies(w, h.authCfg)
	response.NoContent(w)
}

// Me handles GET /api/v1/auth/me (authenticated).
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	user, err := h.svc.Me(r.Context(), identity.UserID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, userEnvelope{User: user}, "")
}

// VerifyEmail handles POST /api/v1/auth/verify-email.
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.svc.VerifyEmail(r.Context(), req.Token); err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, nil, "Email verified")
}

// ForgotPassword handles POST /api/v1/auth/forgot-password. The response is
// intentionally generic to avoid revealing whether an account exists.
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.svc.ForgotPassword(r.Context(), req.Email); err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, nil, "If an account exists, a password reset email has been sent.")
}

// ResetPassword handles POST /api/v1/auth/reset-password.
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.svc.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		response.Error(w, r, err)
		return
	}
	clearAuthCookies(w, h.authCfg)
	response.OK(w, nil, "Password updated. Please sign in again.")
}

func (h *Handler) writeSessionCookies(w http.ResponseWriter, result LoginResult) {
	setAuthCookies(w, h.authCfg, result.AccessToken, h.accessTTL, result.RefreshToken, h.refreshTTL)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return false
	}
	return true
}

func clientMeta(r *http.Request) ClientMeta {
	return ClientMeta{IPAddress: middleware.ClientIP(r), UserAgent: r.UserAgent()}
}

func refreshTokenFromCookie(r *http.Request) string {
	if cookie, err := r.Cookie(refreshCookieName); err == nil {
		return cookie.Value
	}
	return ""
}
