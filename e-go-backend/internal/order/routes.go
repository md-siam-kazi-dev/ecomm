package order

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/internal/audit"
	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/internal/config"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
	"github.com/md-siam-kazi/e-go-backend/pkg/middleware"
	"github.com/md-siam-kazi/e-go-backend/pkg/payment"
	"github.com/md-siam-kazi/e-go-backend/pkg/ratelimit"
)

// Deps is the dependency bundle for the order module.
type Deps struct {
	DB          *pgxpool.Pool
	Cfg         *config.Config
	Registry    *payment.Registry
	FrontendURL string
	RateLimiter ratelimit.Limiter
}

// BuildService constructs the module service from configuration.
func BuildService(db *pgxpool.Pool, cfg *config.Config, registry *payment.Registry, frontendURL string) *Service {
	return NewService(NewRepository(db), registry, Config{
		Currency:              cfg.Payment.Currency,
		StandardShipping:      cfg.Order.StandardShipping,
		ExpressShipping:       cfg.Order.ExpressShipping,
		FreeShippingThreshold: cfg.Order.FreeShippingThreshold,
		ProviderName:          cfg.Payment.Provider,
		SuccessURL:            frontendURL + "/order/success",
		CancelURL:             frontendURL + "/checkout",
	})
}

// Register mounts the order routes: an authenticated group for create/list/get
// and a public provider webhook.
func Register(api chi.Router, signer *jwt.Signer, deps Deps) {
	h := NewHandler(BuildService(deps.DB, deps.Cfg, deps.Registry, deps.FrontendURL))
	h.audit = audit.NewRecorder(deps.DB)

	// Public: verified server-to-server provider callback.
	api.Post("/orders/webhook/{provider}", h.Webhook)

	api.Group(func(r chi.Router) {
		r.Use(auth.Authenticate(signer))

		r.With(middleware.RateLimit(deps.RateLimiter, middleware.RateKey("order_create"), 30, time.Hour)).
			Post("/orders", h.CreateOrder)
		r.Get("/orders", h.ListOrders)
		r.Get("/orders/{id}", h.GetOrder)
	})
}
