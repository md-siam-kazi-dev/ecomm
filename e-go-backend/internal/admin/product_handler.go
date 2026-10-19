package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

var errInvalidBody = apperr.New(http.StatusBadRequest, "BAD_REQUEST", "Request body must be valid JSON")

// ProductHandler handles admin product moderation.
type ProductHandler struct {
	svc *ProductService
}

func NewProductHandler(svc *ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// ListProducts handles GET /api/v1/admin/products?status=&search=&page=&limit=.
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	items, pagination, err := h.svc.List(r.Context(), ReviewQuery{
		Status: q.Get("status"), Search: q.Get("search"), Page: page, Limit: limit,
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// GetProduct handles GET /api/v1/admin/products/{id}.
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	product, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, product, "")
}

// Approve handles POST /api/v1/admin/products/{id}/approve.
func (h *ProductHandler) Approve(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	if err := h.svc.Approve(r.Context(), chi.URLParam(r, "id"), identity.UserID); err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, nil, "Product approved")
}

// Reject handles POST /api/v1/admin/products/{id}/reject.
func (h *ProductHandler) Reject(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req ReviewDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	if err := h.svc.Reject(r.Context(), identity.UserID, chi.URLParam(r, "id"), req.Reason); err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, nil, "Product rejected")
}

// RequestChanges handles POST /api/v1/admin/products/{id}/changes.
func (h *ProductHandler) RequestChanges(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req ReviewDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	if err := h.svc.RequestChanges(r.Context(), identity.UserID, chi.URLParam(r, "id"), req.Note); err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, nil, "Changes requested")
}
