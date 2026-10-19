package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

type ctxKey int

const identityKey ctxKey = iota

// Identity is the authenticated caller extracted from an access token.
type Identity struct {
	UserID string
	Role   string
}

// Authenticate validates the access token (Authorization: Bearer or the
// access_token cookie) and stores the caller identity in the request context.
// Any failure is a 401. Handlers behind this middleware can assume an identity.
func Authenticate(signer *jwt.Signer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := accessToken(r)
			if token == "" {
				response.Error(w, r, apperr.ErrUnauthorized)
				return
			}

			claims, err := signer.Parse(token)
			if err != nil {
				response.Error(w, r, apperr.ErrUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), identityKey, Identity{
				UserID: claims.Subject,
				Role:   claims.Role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthenticate is like Authenticate but never rejects: a missing or
// invalid token leaves the request unauthenticated. Use it on public endpoints
// that enrich their response when the caller happens to be signed in.
func OptionalAuthenticate(signer *jwt.Signer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := accessToken(r)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := signer.Parse(token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), identityKey, Identity{
				UserID: claims.Subject,
				Role:   claims.Role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RoleChecker resolves whether a user currently holds any of the given roles.
// It is implemented against the database so authorization is authoritative:
// role changes take effect immediately and a user holding several roles (for
// example admin *and* vendor) is recognized for each of them.
type RoleChecker interface {
	HasAnyRole(ctx context.Context, userID string, roles ...string) (bool, error)
}

// RequireRole allows only users who currently hold one of the listed roles.
// It must run after Authenticate: a missing identity is a 401, an authenticated
// user without any of the roles is a 403. On a checker error the request is
// denied (fail closed).
func RequireRole(checker RoleChecker, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := IdentityFromContext(r.Context())
			if !ok {
				response.Error(w, r, apperr.ErrUnauthorized)
				return
			}

			allowed, err := checker.HasAnyRole(r.Context(), identity.UserID, roles...)
			if err != nil {
				slog.ErrorContext(r.Context(), "role check failed",
					slog.String("error", err.Error()),
					slog.String("path", r.URL.Path),
				)
				response.Error(w, r, apperr.ErrInternal.WithCause(err))
				return
			}
			if !allowed {
				response.Error(w, r, apperr.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// IdentityFromContext returns the caller identity stored by Authenticate.
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(identityKey).(Identity)
	return identity, ok
}

// accessToken reads the caller's access token from the Authorization header,
// falling back to the access_token cookie.
func accessToken(r *http.Request) string {
	if token := bearerToken(r); token != "" {
		return token
	}
	if cookie, err := r.Cookie(accessCookieName); err == nil {
		return cookie.Value
	}
	return ""
}

func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
