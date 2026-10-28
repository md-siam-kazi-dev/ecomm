package security

import (
	"context"
	"net/netip"
	"strings"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

const (
	defaultPageLimit = 25
	// defaultListWindow keeps the history bounded when the caller sends no range,
	// so an admin page cannot accidentally scan the whole retention window.
	defaultListWindow = 24 * time.Hour
	// maxListWindow matches the documented retention ceiling.
	maxListWindow = 7 * 24 * time.Hour
	maxSearchText = 200
)

var allowedLimits = []int{10, 25, 50, 100}

var allowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

// Service is the read side of security monitoring.
type Service struct {
	repo      *Repository
	collector *Collector
	now       func() time.Time
}

// NewService builds the read service.
func NewService(repo *Repository, collector *Collector) *Service {
	return &Service{repo: repo, collector: collector, now: time.Now}
}

// Stats returns the live overview aggregates. These come from the in-memory
// collector, never from the events loaded in a browser.
func (s *Service) Stats(r RangeName) Snapshot {
	if !r.Valid() {
		r = Range15m
	}
	return s.collector.Stats(r)
}

// Dropped reports telemetry shed by the collector.
func (s *Service) Dropped() int64 { return s.collector.DroppedEvents() }

// List returns a page of persisted events after normalizing the filters.
func (s *Service) List(ctx context.Context, q ListQuery) ([]EventRecord, response.Pagination, error) {
	q = s.normalize(q)

	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, response.Pagination{}, err
	}

	return items, response.Pagination{
		Page:  q.Page,
		Limit: q.Limit,
		Total: total,
	}, nil
}

// Get returns one event.
func (s *Service) Get(ctx context.Context, id string) (EventRecord, error) {
	if !validUUID(id) {
		return EventRecord{}, errInvalidID
	}
	return s.repo.Get(ctx, id)
}

var errInvalidID = apperr.New(400, "INVALID_ID", "Invalid event id")

// normalize applies allow-lists and range bounds. Unknown filter values are
// dropped rather than rejected, matching the audit trail's behaviour: a stale
// bookmark must not 4xx the whole page.
func (s *Service) normalize(q ListQuery) ListQuery {
	if q.Page < 1 {
		q.Page = 1
	}
	if !containsInt(allowedLimits, q.Limit) {
		q.Limit = defaultPageLimit
	}

	q.Method = strings.ToUpper(strings.TrimSpace(q.Method))
	if !containsString(allowedMethods, q.Method) {
		q.Method = ""
	}

	q.EventType = strings.TrimSpace(q.EventType)
	if !SecurityEventType(q.EventType).Valid() {
		q.EventType = ""
	}

	q.Severity = strings.TrimSpace(strings.ToLower(q.Severity))
	if !SecuritySeverity(q.Severity).Valid() {
		q.Severity = ""
	}

	q.Endpoint = truncateString(strings.TrimSpace(q.Endpoint), maxSearchText)

	// An invalid IP or user id would fail the ::inet / ::uuid cast, so drop it.
	if !validIP(q.IP) {
		q.IP = ""
	}
	if q.UserID != "" && !validUUID(q.UserID) {
		q.UserID = ""
	}

	now := s.now()
	if q.To.IsZero() {
		q.To = now
	}
	if q.From.IsZero() {
		q.From = q.To.Add(-defaultListWindow)
	}
	if q.From.After(q.To) {
		q.From, q.To = q.To, q.From
	}
	if q.To.Sub(q.From) > maxListWindow {
		q.From = q.To.Add(-maxListWindow)
	}

	return q
}

func containsInt(values []int, needle int) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func containsString(values []string, needle string) bool {
	if needle == "" {
		return false
	}
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

// validIP reports whether the filter can be safely cast to ::inet.
func validIP(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}
	_, err := netip.ParseAddr(value)
	return err == nil
}
