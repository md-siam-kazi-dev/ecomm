package feedback

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
	"github.com/md-siam-kazi/e-go-backend/pkg/middleware"
	"github.com/md-siam-kazi/e-go-backend/pkg/ratelimit"
)

// Register mounts the authenticated customer feedback routes (any role).
func Register(api chi.Router, signer *jwt.Signer, db *pgxpool.Pool, limiter ratelimit.Limiter) {
	mount(api, NewHandler(NewService(NewRepository(db))), signer, limiter)
}

// mount wires the handlers. Split out so tests can inject fakes.
func mount(api chi.Router, h *Handler, signer *jwt.Signer, limiter ratelimit.Limiter) {
	api.Group(func(r chi.Router) {
		r.Use(auth.Authenticate(signer))
		r.With(middleware.RateLimit(limiter, middleware.RateKey("review_create"), 30, time.Hour)).
			Post("/reviews", h.CreateReview)
		r.Put("/reviews/{id}", h.UpdateReview)
		r.With(middleware.RateLimit(limiter, middleware.RateKey("question_create"), 20, time.Hour)).
			Post("/products/{slug}/questions", h.CreateQuestion)
	})
}
