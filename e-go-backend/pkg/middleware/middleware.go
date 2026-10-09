package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/cors"
	"github.com/google/uuid"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/ratelimit"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

var errInternalPanic = apperr.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong")

type ctxKey int

const requestIDKey ctxKey = iota

// RequestID attaches a request ID to the context and echoes it back in the
// X-Request-Id response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-Id", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext returns the request ID stored by RequestID, if any.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// statusRecorder captures the response status for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wrote {
		r.status = code
		r.wrote = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wrote {
		r.status = http.StatusOK
		r.wrote = true
	}
	return r.ResponseWriter.Write(b)
}

// Logger emits one structured log line per request.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			log.InfoContext(r.Context(), "http_request",
				slog.String("request_id", RequestIDFromContext(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			)
		})
	}
}

// Recoverer catches panics, logs them with a stack trace, and returns a
// standardized 500 without leaking internals to the client.
func Recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.ErrorContext(r.Context(), "panic recovered",
						slog.Any("panic", rec),
						slog.String("request_id", RequestIDFromContext(r.Context())),
						slog.String("path", r.URL.Path),
						slog.String("stack", string(debug.Stack())),
					)
					response.Error(w, r, errInternalPanic)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORS applies the configured allow-list. Wildcard origins are intentionally
// not supported for authenticated APIs.
func CORS(origins, methods, headers []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   methods,
		AllowedHeaders:   headers,
		ExposedHeaders:   []string{"X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

// MaxBytes limits the size of request bodies.
func MaxBytes(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

// SecureHeaders sets a minimal set of hardening headers.
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

// ClientIP extracts the best-effort client IP, honouring the usual proxy
// headers before falling back to the socket address.
func ClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if first, _, found := strings.Cut(forwarded, ","); found || first != "" {
			return strings.TrimSpace(first)
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateKey builds a limiter key function scoped to a route name and the client
// IP, e.g. "auth_login:203.0.113.7".
func RateKey(scope string) func(*http.Request) string {
	return func(r *http.Request) string {
		return scope + ":" + ClientIP(r)
	}
}

// RateLimit rejects requests that exceed the fixed-window limit for their key.
// Limiter errors fail open (allow) so a Redis outage cannot lock users out.
func RateLimit(limiter ratelimit.Limiter, keyFn func(*http.Request) string, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter == nil {
				next.ServeHTTP(w, r)
				return
			}

			allowed, err := limiter.Allow(r.Context(), keyFn(r), limit, window)
			if err != nil {
				slog.WarnContext(r.Context(), "rate limiter unavailable; allowing request",
					slog.String("error", err.Error()),
					slog.String("path", r.URL.Path),
				)
				next.ServeHTTP(w, r)
				return
			}
			if !allowed {
				response.Error(w, r, apperr.ErrRateLimited)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CSRFOriginCheck rejects state-changing requests whose Origin header is not on
// the allow-list. It complements SameSite=Lax cookies as the CSRF defence for
// cookie-authenticated endpoints. Requests without an Origin header (e.g. curl,
// same-origin form posts) are allowed.
func CSRFOriginCheck(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[strings.ToLower(strings.TrimRight(origin, "/"))] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
				origin := r.Header.Get("Origin")
				if origin != "" {
					if _, ok := allowed[strings.ToLower(strings.TrimRight(origin, "/"))]; !ok {
						response.Error(w, r, apperr.ErrForbidden)
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
