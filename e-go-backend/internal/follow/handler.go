package follow

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// Handler translates authenticated store-follow requests into service calls.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func identity(r *http.Request) (auth.Identity, bool) {
	return auth.IdentityFromContext(r.Context())
}

// Status handles GET /api/v1/stores/{slug}/follow. Public: guests get
// following=false plus the store's follower count.
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	var userID string
	if id, ok := identity(r); ok {
		userID = id.UserID
	}
	status, err := h.svc.Status(r.Context(), chi.URLParam(r, "slug"), userID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, status, "")
}

// Follow handles POST /api/v1/stores/{slug}/follow.
func (h *Handler) Follow(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}
	status, err := h.svc.Follow(r.Context(), chi.URLParam(r, "slug"), id.UserID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, status, "Following")
}

// Unfollow handles DELETE /api/v1/stores/{slug}/follow.
func (h *Handler) Unfollow(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}
	status, err := h.svc.Unfollow(r.Context(), chi.URLParam(r, "slug"), id.UserID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, status, "Unfollowed")
}
