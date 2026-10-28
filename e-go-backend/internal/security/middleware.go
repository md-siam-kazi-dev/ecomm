package security

import (
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/audit"
	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/middleware"
)

// maxEndpointLength bounds the route-pattern column. Route patterns are short;
// this only matters for unmatched paths, which fall back to the raw URL.
const maxEndpointLength = 200

type ctxKey int

const streamStateKey ctxKey = iota

// streamState lets a handler tell the middleware that this request became a
// long-lived stream, so it is not counted as a request datapoint (its duration
// would be minutes and would wreck the average response time).
type streamState struct {
	streaming atomic.Bool
}

// MarkStreaming flags the current request as a long-lived stream. Called by the
// SSE handler; a no-op outside the monitoring middleware.
func MarkStreaming(ctx context.Context) {
	if state, ok := ctx.Value(streamStateKey).(*streamState); ok {
		state.streaming.Store(true)
	}
}

// Middleware captures safe request metadata and hands it to the collector.
//
// It must run after the request ID and client-IP middleware, and after an
// optional authentication pass so the caller identity is available. It is
// mounted on the versioned API router rather than globally, so the health probes
// are not counted as traffic.
//
// Safety properties:
//   - the wrapped writer forwards Flush and Unwrap, so a stream can still flush
//     and set its own write deadline through it;
//   - collection is wrapped in a recover, so a monitoring bug can never fail a
//     request;
//   - no body, header, cookie or token is ever read.
func Middleware(c *Collector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c == nil {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			state := &streamState{}
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r.WithContext(context.WithValue(r.Context(), streamStateKey, state)))

			if state.streaming.Load() {
				return
			}

			safeCollect(c, r, rec.status, time.Since(start))
		})
	}
}

func safeCollect(c *Collector, r *http.Request, status int, took time.Duration) {
	// Monitoring must never be able to break a request.
	defer func() { _ = recover() }()

	c.RecordRequest(r.Context(), buildEvent(r, status, took, c.clock()()))
}

func buildEvent(r *http.Request, status int, took time.Duration, at time.Time) SecurityEvent {
	ev := SecurityEvent{
		RequestID:  middleware.RequestIDFromContext(r.Context()),
		IP:         middleware.ClientIP(r),
		Method:     r.Method,
		Endpoint:   sanitizeEndpoint(routePattern(r)),
		StatusCode: status,
		DurationMS: int(took.Milliseconds()),
		UserAgent:  truncateString(r.UserAgent(), maxUserAgent),
		CreatedAt:  at,
	}

	// The audit middleware already resolved the same values; prefer them so both
	// trails describe the request identically.
	if info, ok := audit.RequestInfoFromContext(r.Context()); ok {
		if ev.IP == "" {
			ev.IP = info.IP
		}
		if ev.RequestID == "" {
			ev.RequestID = info.RequestID
		}
	}

	if identity, ok := auth.IdentityFromContext(r.Context()); ok {
		ev.UserID = identity.UserID
		ev.UserRole = identity.Role
	}

	if eventType, severity, classified := ClassifyHTTP(status); classified {
		ev.Type = eventType
		ev.Severity = severity
	}

	return ev
}

// routePattern returns the matched route template (e.g. /api/v1/products/:id)
// so the monitoring table never records one row per product id. Unmatched
// requests keep their path, sanitized.
func routePattern(r *http.Request) string {
	if rc := chi.RouteContext(r.Context()); rc != nil {
		if pattern := rc.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	return r.URL.Path
}

// sanitizeEndpoint strips control characters (which would allow log/HTML
// injection) and bounds the length.
func sanitizeEndpoint(value string) string {
	if value == "" {
		return ""
	}
	cleaned := strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, value)
	if len(cleaned) > maxEndpointLength {
		cleaned = cleaned[:maxEndpointLength]
	}
	return cleaned
}

// responseRecorder captures the status code while staying transparent to
// streaming handlers.
type responseRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.wrote {
		r.status = code
		r.wrote = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wrote {
		r.status = http.StatusOK
		r.wrote = true
	}
	return r.ResponseWriter.Write(b)
}

// Flush forwards to the underlying writer so Server-Sent Events keep streaming.
func (r *responseRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Unwrap lets http.ResponseController reach the real writer, which the SSE
// handler needs to clear the server's write deadline.
func (r *responseRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }
