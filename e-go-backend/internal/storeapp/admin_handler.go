package storeapp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// AdminHandler serves the admin store-application review endpoints. The route
// group is already wrapped by auth.Authenticate + RequireRole("admin").
type AdminHandler struct {
	svc *Service
}

func NewAdminHandler(svc *Service) *AdminHandler { return &AdminHandler{svc: svc} }

func adminQueryFrom(r *http.Request) AdminListQuery {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	return AdminListQuery{
		Page:          page,
		Limit:         limit,
		Status:        q.Get("status"),
		PaymentStatus: q.Get("payment_status"),
		Search:        q.Get("search"),
		Sort:          q.Get("sort"),
	}
}

// ListApplications handles GET /api/v1/admin/store-applications.
func (h *AdminHandler) ListApplications(w http.ResponseWriter, r *http.Request) {
	items, pagination, err := h.svc.AdminListApplications(r.Context(), adminQueryFrom(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// GetApplication handles GET /api/v1/admin/store-applications/{id}.
func (h *AdminHandler) GetApplication(w http.ResponseWriter, r *http.Request) {
	detail, err := h.svc.AdminDetail(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, detail, "")
}

// Approve handles POST /api/v1/admin/store-applications/{id}/approve.
func (h *AdminHandler) Approve(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	detail, err := h.svc.Approve(r.Context(), identity.UserID, chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, detail, "Application approved")
}

// Reject handles POST /api/v1/admin/store-applications/{id}/reject.
func (h *AdminHandler) Reject(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req RejectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}
	if err := h.svc.Reject(r.Context(), identity.UserID, chi.URLParam(r, "id"), req.Reason); err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, map[string]string{"status": StatusRejected}, "Application rejected")
}

// UpdateNotes handles POST /api/v1/admin/store-applications/{id}/notes.
func (h *AdminHandler) UpdateNotes(w http.ResponseWriter, r *http.Request) {
	var req NoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}
	if err := h.svc.UpdateNotes(r.Context(), chi.URLParam(r, "id"), req.Note); err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, map[string]string{"saved": "true"}, "Note saved")
}

// ListStores handles GET /api/v1/admin/stores.
func (h *AdminHandler) ListStores(w http.ResponseWriter, r *http.Request) {
	items, pagination, err := h.svc.AdminListStores(r.Context(), adminQueryFrom(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// ListPayments handles GET /api/v1/admin/application-payments.
func (h *AdminHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	items, pagination, err := h.svc.AdminListPayments(r.Context(), adminQueryFrom(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}
