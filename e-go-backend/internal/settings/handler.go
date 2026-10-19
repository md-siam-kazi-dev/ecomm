package settings

import (
	"encoding/json"
	"net/http"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// Handler exposes the storefront settings API.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Get handles GET /api/v1/storefront/settings and GET /api/v1/admin/settings.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.svc.Get(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, item, "")
}

// Update handles PUT /api/v1/admin/settings.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	item, err := h.svc.Update(r.Context(), identity.UserID, req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, item, "Storefront settings updated")
}
