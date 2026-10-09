package admin

import (
	"context"
	"strings"

	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

const (
	defaultLimit = 20
	maxSearchLen = 100
)

// allowedLimits is the set of page sizes the API accepts; anything else falls
// back to defaultLimit.
var allowedLimits = map[int]struct{}{10: {}, 20: {}, 50: {}, 100: {}}

// allowedRoles are the role filter values; empty means "all".
var allowedRoles = map[string]struct{}{"user": {}, "vendor": {}, "admin": {}}

// Service owns the admin business logic (Handler → Service → Repository).
type Service struct {
	users UserRepository
}

func NewService(users UserRepository) *Service { return &Service{users: users} }

// ListUsers normalizes the query, then returns one page of users plus the
// pagination metadata. TotalPages/HasNext/HasPrevious are derived by the
// response layer.
func (s *Service) ListUsers(ctx context.Context, q ListUsersQuery) ([]UserItem, response.Pagination, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}

	limit := q.Limit
	if _, ok := allowedLimits[limit]; !ok {
		limit = defaultLimit
	}

	role := strings.ToLower(strings.TrimSpace(q.Role))
	if role == "customer" { // friendly alias from the UI
		role = "user"
	}
	if _, ok := allowedRoles[role]; !ok {
		role = ""
	}

	search := strings.TrimSpace(q.Search)
	if len(search) > maxSearchLen {
		search = search[:maxSearchLen]
	}

	records, total, err := s.users.ListUsers(ctx, ListUsersQuery{
		Page:   page,
		Limit:  limit,
		Role:   role,
		Search: search,
	})
	if err != nil {
		return nil, response.Pagination{}, err
	}

	items := make([]UserItem, 0, len(records))
	for _, rec := range records {
		items = append(items, toUserItem(rec))
	}

	return items, response.Pagination{Page: page, Limit: limit, Total: total}, nil
}
