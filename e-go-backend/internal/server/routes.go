package server

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/account"
	"github.com/md-siam-kazi/e-go-backend/internal/admin"
	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/internal/catalog"
	"github.com/md-siam-kazi/e-go-backend/internal/database"
	"github.com/md-siam-kazi/e-go-backend/internal/feedback"
	"github.com/md-siam-kazi/e-go-backend/internal/order"
	"github.com/md-siam-kazi/e-go-backend/internal/storeapp"
	"github.com/md-siam-kazi/e-go-backend/internal/vendor"
	"github.com/md-siam-kazi/e-go-backend/pkg/imgbb"
	"github.com/md-siam-kazi/e-go-backend/pkg/middleware"
	"github.com/md-siam-kazi/e-go-backend/pkg/payment"
	"github.com/md-siam-kazi/e-go-backend/pkg/ratelimit"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// buildRouter assembles the global middleware stack and mounts all route
// groups. New domain modules register their routes inside the /api/v1 group.
func (s *Server) buildRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(s.log))
	r.Use(middleware.Recoverer(s.log))
	r.Use(middleware.SecureHeaders)
	r.Use(middleware.CORS(
		s.cfg.CORS.AllowedOrigins,
		s.cfg.CORS.AllowedMethods,
		s.cfg.CORS.AllowedHeaders,
	))
	r.Use(middleware.MaxBytes(s.cfg.App.MaxRequestBodyBytes))
	// CSRF defence for cookie-authenticated writes: reject state-changing
	// requests whose Origin is not an allowed frontend origin.
	r.Use(middleware.CSRFOriginCheck(s.cfg.CORS.AllowedOrigins))

	// Infrastructure probes are unauthenticated and outside the API version.
	r.Get("/health", s.handleHealth)
	r.Get("/ready", s.handleReady)

	r.Route("/api/v1", func(api chi.Router) {
		s.registerModules(api)
	})

	return r
}

// registerModules is the single place where domain modules wire their routes.
// Each module exposes its own routes.go with a Register(router) function.
func (s *Server) registerModules(api chi.Router) {
	auth.Register(api, auth.Deps{
		DB: s.db, Redis: s.redis, Cfg: s.cfg, Log: s.log, Signer: s.signer,
	})

	// Public storefront catalog (read-only, no auth).
	catalog.Register(api, s.db)

	// Role checks are authoritative (read from user_roles), so a role change
	// takes effect immediately and a user holding several roles is recognized
	// for each of them.
	roleChecker := auth.NewRepository(s.db)

	// Store applications / vendor onboarding. Applicant routes are auth-only;
	// admin review routes mount inside the /admin group below.
	frontendURL := strings.TrimRight(s.cfg.Auth.FrontendURL, "/")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	paymentRegistry := payment.NewRegistryFromConfig(payment.Config{
		Provider:      s.cfg.Payment.Provider,
		SecretKey:     s.cfg.Payment.ProviderKey,
		WebhookSecret: s.cfg.Payment.WebhookSecret,
		Currency:      s.cfg.Payment.Currency,
		SuccessURL:    frontendURL + "/order/success",
		CancelURL:     frontendURL + "/checkout",
	})
	storeappSvc := storeapp.BuildService(s.db, s.cfg, paymentRegistry)
	storeapp.RegisterApplicant(api, s.signer, storeappSvc, ratelimit.New(s.redis))

	// Storefront orders + Stripe checkout.
	order.Register(api, s.signer, order.Deps{
		DB:          s.db,
		Cfg:         s.cfg,
		Registry:    paymentRegistry,
		FrontendURL: frontendURL,
		RateLimiter: ratelimit.New(s.redis),
	})

	// Customer feedback (reviews + product questions) — authenticated writes.
	feedback.Register(api, s.signer, s.db, ratelimit.New(s.redis))

	// Admin-only area: authenticate, then require the admin role.
	api.Route("/admin", func(r chi.Router) {
		r.Use(auth.Authenticate(s.signer))
		r.Use(auth.RequireRole(roleChecker, "admin"))
		admin.Register(r, s.db)
		storeapp.RegisterAdmin(r, storeappSvc)
	})

	// Vendor-only area: authenticate, then require the vendor role.
	api.Route("/vendor", func(r chi.Router) {
		r.Use(auth.Authenticate(s.signer))
		r.Use(auth.RequireRole(roleChecker, "vendor"))
		vendor.Register(r, s.db, imgbb.New(s.cfg.Images.ImgBBAPIKey))
	})

	// Any authenticated user: their own account data (profile, addresses, ...).
	api.Route("/account", func(r chi.Router) {
		r.Use(auth.Authenticate(s.signer))
		account.Register(r, s.db, imgbb.New(s.cfg.Images.ImgBBAPIKey))
	})

	// Remaining modules (products, carts, orders, payments, ...)
	// will be mounted here as they are implemented.
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, map[string]string{
		"status":  "ok",
		"service": "e-go-backend",
		"env":     s.cfg.App.Env,
	}, "")
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	deps := map[string]string{
		"postgres": "ok",
		"redis":    "ok",
	}

	ready := true

	if err := database.PingPostgres(ctx, s.db); err != nil {
		deps["postgres"] = "unavailable"
		ready = false
	}
	if err := database.PingRedis(ctx, s.redis); err != nil {
		deps["redis"] = "unavailable"
		ready = false
	}

	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}
	response.JSON(w, status, map[string]interface{}{
		"ready":        ready,
		"dependencies": deps,
	}, "")
}
