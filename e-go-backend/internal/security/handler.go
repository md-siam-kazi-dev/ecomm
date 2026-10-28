package security

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// Handler exposes the read API and the live stream.
type Handler struct {
	service  *Service
	streamer *Streamer
}

// NewHandler builds the admin handler.
func NewHandler(service *Service, streamer *Streamer) *Handler {
	return &Handler{service: service, streamer: streamer}
}

// Stats returns the aggregated overview for one range.
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	rangeName := RangeName(strings.TrimSpace(r.URL.Query().Get("range")))
	if !rangeName.Valid() {
		rangeName = Range15m
	}

	response.OK(w, statsResponse{
		Snapshot:      h.service.Stats(rangeName),
		DroppedEvents: h.service.Dropped(),
	}, "")
}

// List returns a page of persisted events.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, pagination, err := h.service.List(r.Context(), queryFrom(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// Get returns one event by id.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	record, err := h.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, record, "")
}

// Stream is the authenticated Server-Sent Events endpoint.
func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	h.streamer.Serve(w, r)
}

func queryFrom(r *http.Request) ListQuery {
	values := r.URL.Query()

	q := ListQuery{
		Page:      atoiOr(values.Get("page"), 1),
		Limit:     atoiOr(values.Get("limit"), defaultPageLimit),
		Method:    values.Get("method"),
		Status:    atoiOr(values.Get("status"), 0),
		IP:        values.Get("ip"),
		UserID:    values.Get("user_id"),
		Endpoint:  values.Get("endpoint"),
		EventType: values.Get("event_type"),
		Severity:  values.Get("severity"),
	}

	if from, ok := parseTime(values.Get("from")); ok {
		q.From = from
	}
	if to, ok := parseTime(values.Get("to")); ok {
		q.To = to
	}
	return q
}

func atoiOr(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

// parseTime accepts RFC3339 or a unix timestamp in seconds.
func parseTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC(), true
	}
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Unix(seconds, 0).UTC(), true
	}
	return time.Time{}, false
}
