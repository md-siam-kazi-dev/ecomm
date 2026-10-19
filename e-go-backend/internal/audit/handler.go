package audit

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

var (
	errInvalidID   = apperr.New(http.StatusBadRequest, "INVALID_ID", "Audit event id must be a valid UUID")
	errInvalidDate = apperr.New(http.StatusBadRequest, "INVALID_DATE", "Dates must be YYYY-MM-DD or RFC3339")
)

// Handler exposes the read-only audit API to admins.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List handles GET /api/v1/admin/audit-logs.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q, err := queryFrom(r)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	items, pagination, err := h.svc.List(r.Context(), q)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// Summary handles GET /api/v1/admin/audit-logs/summary.
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	q, err := queryFrom(r)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	summary, err := h.svc.Summary(r.Context(), q)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, summary, "")
}

// Get handles GET /api/v1/admin/audit-logs/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	event, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, event, "")
}

// queryFrom parses and validates the list/summary query string.
func queryFrom(r *http.Request) (ListQuery, error) {
	values := r.URL.Query()
	page, _ := strconv.Atoi(values.Get("page"))
	limit, _ := strconv.Atoi(values.Get("limit"))

	q := ListQuery{
		Page:       page,
		Limit:      limit,
		Search:     values.Get("search"),
		Action:     values.Get("action"),
		ActorRole:  values.Get("actor_role"),
		EntityType: values.Get("entity_type"),
		EntityID:   values.Get("entity_id"),
		Result:     values.Get("result"),
	}

	from, err := parseDate(values.Get("from"), false)
	if err != nil {
		return ListQuery{}, err
	}
	to, err := parseDate(values.Get("to"), true)
	if err != nil {
		return ListQuery{}, err
	}
	q.From, q.To = from, to
	return q, nil
}

// parseDate accepts YYYY-MM-DD (or RFC3339). For an inclusive "to" date, the
// end of the day is used so events on that date are included.
func parseDate(v string, endOfDay bool) (*time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return &t, nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil, errInvalidDate
	}
	if endOfDay {
		t = t.AddDate(0, 0, 1)
	}
	return &t, nil
}
