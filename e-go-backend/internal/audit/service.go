package audit

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

const (
	defaultLimit = 25
	maxSearchLen = 200
	maxRangeDays = 366
)

var allowedLimits = map[int]struct{}{10: {}, 25: {}, 50: {}, 100: {}}

var allowedRoles = map[string]struct{}{
	RoleAdmin: {}, RoleVendor: {}, RoleCustomer: {}, RoleSystem: {},
}

var allowedResults = map[string]struct{}{ResultSuccess: {}, ResultFailure: {}}

var allowedEntityTypes = map[string]struct{}{
	EntityUser: {}, EntityStore: {}, EntityStoreApplication: {}, EntityProduct: {},
	EntityVariant: {}, EntityOrder: {}, EntityPayment: {}, EntityReview: {},
	EntityQuestion: {}, EntityInventory: {}, EntitySettings: {},
}

// Service owns the audit read logic. Audit rows are append-only: there is no
// create/update/delete exposed here.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

// List normalizes the filter and returns one page of events plus pagination.
func (s *Service) List(ctx context.Context, q ListQuery) ([]EventItem, response.Pagination, error) {
	q = normalize(q)
	records, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, response.Pagination{}, err
	}
	items := make([]EventItem, 0, len(records))
	for _, rec := range records {
		items = append(items, toEventItem(rec))
	}
	return items, response.Pagination{Page: q.Page, Limit: q.Limit, Total: total}, nil
}

// Get returns the full detail of a single event.
func (s *Service) Get(ctx context.Context, id string) (EventDetail, error) {
	if _, err := uuid.Parse(id); err != nil {
		return EventDetail{}, errInvalidID
	}
	rec, err := s.repo.Get(ctx, id)
	if err != nil {
		return EventDetail{}, err
	}
	return EventDetail{
		EventItem: toEventItem(rec.EventRecord),
		OldValues: rec.OldValues,
		NewValues: rec.NewValues,
		Metadata:  rec.Metadata,
		IPAddress: rec.IPAddress,
		UserAgent: rec.UserAgent,
		RequestID: rec.RequestID,
	}, nil
}

// Summary aggregates the selected window and returns the action catalogue.
func (s *Service) Summary(ctx context.Context, q ListQuery) (Summary, error) {
	q = normalize(q)
	counts, err := s.repo.Summary(ctx, q)
	if err != nil {
		return Summary{}, err
	}
	return Summary{
		Total:          counts.Total,
		AdminActions:   counts.AdminActions,
		VendorActions:  counts.VendorActions,
		SecurityEvents: counts.SecurityEvents,
		From:           formatTime(q.From),
		To:             formatTime(q.To),
		Actions:        Actions(),
	}, nil
}

// normalize applies defaults, bounds and allow-lists to a raw query.
func normalize(q ListQuery) ListQuery {
	if q.Page < 1 {
		q.Page = 1
	}
	if _, ok := allowedLimits[q.Limit]; !ok {
		q.Limit = defaultLimit
	}
	q.Search = strings.TrimSpace(q.Search)
	if len(q.Search) > maxSearchLen {
		q.Search = q.Search[:maxSearchLen]
	}
	q.Action = strings.TrimSpace(q.Action)
	q.EntityType = strings.ToLower(strings.TrimSpace(q.EntityType))
	if _, ok := allowedEntityTypes[q.EntityType]; !ok {
		q.EntityType = ""
	}
	q.EntityID = strings.TrimSpace(q.EntityID)
	role := strings.ToLower(strings.TrimSpace(q.ActorRole))
	if _, ok := allowedRoles[role]; !ok {
		role = ""
	}
	q.ActorRole = role
	result := strings.ToLower(strings.TrimSpace(q.Result))
	if _, ok := allowedResults[result]; !ok {
		result = ""
	}
	q.Result = result
	q.From, q.To = boundRange(q.From, q.To)
	return q
}

// boundRange clamps an unbounded or over-long window so the summary/list can
// never scan the entire table by accident.
func boundRange(from, to *time.Time) (*time.Time, *time.Time) {
	if from == nil && to == nil {
		return nil, nil
	}
	if from == nil {
		start := to.AddDate(0, 0, -maxRangeDays)
		from = &start
	}
	if to == nil {
		end := from.AddDate(0, 0, maxRangeDays)
		to = &end
	}
	if to.Sub(*from) > time.Duration(maxRangeDays)*24*time.Hour {
		clamped := from.AddDate(0, 0, maxRangeDays)
		to = &clamped
	}
	if to.Before(*from) {
		from, to = to, from
	}
	return from, to
}

func formatTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}
