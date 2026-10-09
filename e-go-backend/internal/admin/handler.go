package admin

import (
	"net/http"
	"strconv"

	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// Handler translates admin HTTP requests into service calls.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ListUsers handles GET /api/v1/admin/users?page=&limit=&role=&search=.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	items, pagination, err := h.svc.ListUsers(r.Context(), ListUsersQuery{
		Page:   page,
		Limit:  limit,
		Role:   query.Get("role"),
		Search: query.Get("search"),
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.List(w, items, pagination)
}
