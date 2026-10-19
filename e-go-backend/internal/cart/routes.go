package cart

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/internal/config"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
	"github.com/md-siam-kazi/e-go-backend/pkg/middleware"
	"github.com/md-siam-kazi/e-go-backend/pkg/ratelimit"
)

// Deps is the dependency bundle for the cart module.
type Deps struct {
	DB          *pgxpool.Pool
	Cfg         *config.Config
	Signer      *jwt.Signer
	RateLimiter ratelimit.Limiter
}

// Register mounts the cart routes. Every route accepts guests and signed-in
// users alike (the middleware never rejects); a signed-in caller uses their
// account cart, a guest uses the cart-session cookie.
func Register(api chi.Router, deps Deps) {
	h := NewHandler(
		NewService(NewRepository(deps.DB), deps.Cfg.Payment.Currency),
		deps.Cfg.Auth,
	)
	mount(api, h, deps.Signer, deps.RateLimiter)
}

// mount wires the handlers. Split out so tests can inject fakes.
func mount(api chi.Router, h *Handler, signer *jwt.Signer, limiter ratelimit.Limiter) {
	api.Group(func(r chi.Router) {
		r.Use(auth.OptionalAuthenticate(signer))

		r.Get("/cart", h.Get)
		r.With(middleware.RateLimit(limiter, middleware.RateKey("cart_add"), 120, time.Hour)).
			Post("/cart/items", h.Add)
		r.Put("/cart/items/{itemId}", h.SetQuantity)
		r.Delete("/cart/items/{itemId}", h.Remove)
		r.Delete("/cart", h.Clear)
	})
}
