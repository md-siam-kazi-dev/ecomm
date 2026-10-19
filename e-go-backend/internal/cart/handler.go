package cart

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/internal/config"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// Handler translates cart HTTP requests into service calls.
type Handler struct {
	svc *Service
	cfg config.AuthConfig
}

func NewHandler(svc *Service, cfg config.AuthConfig) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

// Get handles GET /api/v1/cart.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	out, res, err := h.svc.Get(r.Context(), actorID(r), h.readSession(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	h.respond(w, res, out, "")
}

// Add handles POST /api/v1/cart/items.
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}
	out, res, err := h.svc.Add(r.Context(), actorID(r), h.readSession(r), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	h.respond(w, res, out, "Item added to cart")
}

// SetQuantity handles PUT /api/v1/cart/items/{itemId}.
func (h *Handler) SetQuantity(w http.ResponseWriter, r *http.Request) {
	var req UpdateQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}
	out, res, err := h.svc.SetQuantity(r.Context(), actorID(r), h.readSession(r), chi.URLParam(r, "itemId"), req.Quantity)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	h.respond(w, res, out, "")
}

// Remove handles DELETE /api/v1/cart/items/{itemId}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	out, res, err := h.svc.Remove(r.Context(), actorID(r), h.readSession(r), chi.URLParam(r, "itemId"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	h.respond(w, res, out, "Item removed from cart")
}

// Clear handles DELETE /api/v1/cart.
func (h *Handler) Clear(w http.ResponseWriter, r *http.Request) {
	out, res, err := h.svc.Clear(r.Context(), actorID(r), h.readSession(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	h.respond(w, res, out, "Cart cleared")
}

// respond writes the cart and reconciles the guest session cookie.
func (h *Handler) respond(w http.ResponseWriter, res Resolution, out CartResponse, message string) {
	switch {
	case res.NewSession != "":
		h.setSessionCookie(w, res.NewSession, int(SessionTTL.Seconds()))
	case res.ClearSession:
		h.setSessionCookie(w, "", -1)
	}
	response.OK(w, out, message)
}

func (h *Handler) readSession(r *http.Request) string {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

// setSessionCookie writes (maxAge > 0) or expires (maxAge < 0) the guest cart
// cookie, mirroring the auth cookie attributes.
func (h *Handler) setSessionCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		Domain:   h.cfg.CookieDomain,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: sameSiteMode(h.cfg.CookieSameSite),
	})
}

func sameSiteMode(v string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// actorID returns the authenticated user id, or "" for a guest.
func actorID(r *http.Request) string {
	if identity, ok := auth.IdentityFromContext(r.Context()); ok {
		return identity.UserID
	}
	return ""
}
