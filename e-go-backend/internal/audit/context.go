package audit

import (
	"context"
	"net/http"

	"github.com/md-siam-kazi/e-go-backend/pkg/middleware"
)

type ctxKey int

const requestInfoKey ctxKey = iota

// RequestInfo carries the HTTP metadata attached to every audited event so
// repositories and handlers do not need access to the *http.Request.
type RequestInfo struct {
	IP        string
	UserAgent string
	RequestID string
}

// Middleware captures request metadata into the context for later auditing.
// It must run after middleware.RequestID so the correlation id is available.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := RequestInfo{
			IP:        middleware.ClientIP(r),
			UserAgent: truncate(r.UserAgent(), 512),
			RequestID: middleware.RequestIDFromContext(r.Context()),
		}
		ctx := context.WithValue(r.Context(), requestInfoKey, info)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestInfoFromContext returns the captured request metadata, if any.
func RequestInfoFromContext(ctx context.Context) (RequestInfo, bool) {
	info, ok := ctx.Value(requestInfoKey).(RequestInfo)
	return info, ok
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
