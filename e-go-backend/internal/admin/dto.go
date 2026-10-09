package admin

import "time"

// ListUsersQuery is the normalized filter for the admin user list.
type ListUsersQuery struct {
	Page   int
	Limit  int
	Role   string // "", "user", "vendor", "admin"
	Search string
}

// UserItem is a single row in the admin user list. Only fields that are useful
// to an administrator are exposed.
type UserItem struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	AvatarURL     string   `json:"avatar_url"`
	Role          string   `json:"role"` // effective (highest) role
	Roles         []string `json:"roles"`
	Status        string   `json:"status"`
	EmailVerified bool     `json:"email_verified"`
	CreatedAt     string   `json:"created_at"`
	LastLoginAt   *string  `json:"last_login_at"`
}

// primaryRole collapses a user's roles into the most privileged one
// (admin > vendor > user), defaulting to "user".
func primaryRole(roles []string) string {
	rank := func(role string) int {
		switch role {
		case "admin":
			return 3
		case "vendor":
			return 2
		case "user":
			return 1
		default:
			return 0
		}
	}

	best := ""
	for _, role := range roles {
		if rank(role) > rank(best) {
			best = role
		}
	}
	if best == "" {
		return "user"
	}
	return best
}

func toUserItem(rec UserRecord) UserItem {
	item := UserItem{
		ID:            rec.ID,
		Name:          rec.Name,
		Email:         rec.Email,
		AvatarURL:     rec.AvatarURL,
		Role:          primaryRole(rec.Roles),
		Roles:         rec.Roles,
		Status:        rec.Status,
		EmailVerified: rec.EmailVerified,
		CreatedAt:     rec.CreatedAt.UTC().Format(time.RFC3339),
	}
	if rec.LastLoginAt != nil {
		last := rec.LastLoginAt.UTC().Format(time.RFC3339)
		item.LastLoginAt = &last
	}
	return item
}
