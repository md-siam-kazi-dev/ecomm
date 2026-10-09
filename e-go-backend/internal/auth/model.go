package auth

import "time"

// User is the internal representation of an account. It includes
// authentication-sensitive fields (PasswordHash) that must never be serialized
// to a client — responses use UserResponse instead.
type User struct {
	ID            string
	Email         string
	Name          string
	Phone         string
	AvatarURL     string
	Role          string
	Status        string
	EmailVerified bool
	CreatedAt     time.Time
	PasswordHash  string
}

// roleRank orders roles by privilege so an account holding several roles
// resolves to a single effective role (admin > vendor > user).
func roleRank(role string) int {
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

// highestRole returns the most privileged role from a set, defaulting to "user".
func highestRole(roles []string) string {
	best := ""
	for _, role := range roles {
		if roleRank(role) > roleRank(best) {
			best = role
		}
	}
	if best == "" {
		return "user"
	}
	return best
}
