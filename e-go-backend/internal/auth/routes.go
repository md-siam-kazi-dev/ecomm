package auth

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/md-siam-kazi/e-go-backend/internal/config"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
	"github.com/md-siam-kazi/e-go-backend/pkg/middleware"
	"github.com/md-siam-kazi/e-go-backend/pkg/ratelimit"
)

// Deps are the module's collaborators, provided by the server. The signer is
// shared so other modules can protect their routes with the same instance.
type Deps struct {
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Cfg    *config.Config
	Log    *slog.Logger
	Signer *jwt.Signer
}

// Register mounts the authentication routes. Called once from the central
// module registry in internal/server.
func Register(r chi.Router, d Deps) {
	signer := d.Signer

	svc := NewService(ServiceDeps{
		Users:      NewRepository(d.DB),
		Sessions:   NewSessionRepository(d.DB),
		Tokens:     NewTokenRepository(d.DB),
		Signer:     signer,
		Mailer:     NewLogMailer(d.Log, d.Cfg.Auth.FrontendURL, d.Cfg.IsProduction()),
		Log:        d.Log,
		AccessTTL:  d.Cfg.JWT.AccessTTL,
		RefreshTTL: d.Cfg.JWT.RefreshTTL,
	})
	h := NewHandler(svc, d.Cfg.Auth, d.Cfg.JWT.AccessTTL, d.Cfg.JWT.RefreshTTL)

	mount(r, h, signer, ratelimit.New(d.Redis))
}

// mount wires the handlers onto the router. Split out so tests can inject a
// handler built with fakes and a nil limiter.
func mount(r chi.Router, h *Handler, signer *jwt.Signer, limiter ratelimit.Limiter) {
	r.Route("/auth", func(r chi.Router) {
		r.With(middleware.RateLimit(limiter, middleware.RateKey("auth_register"), 10, time.Hour)).
			Post("/register", h.Register)
		r.With(middleware.RateLimit(limiter, middleware.RateKey("auth_login"), 10, 15*time.Minute)).
			Post("/login", h.Login)
		r.With(middleware.RateLimit(limiter, middleware.RateKey("auth_refresh"), 60, 15*time.Minute)).
			Post("/refresh", h.Refresh)
		r.Post("/logout", h.Logout)
		r.With(middleware.RateLimit(limiter, middleware.RateKey("auth_verify"), 20, time.Hour)).
			Post("/verify-email", h.VerifyEmail)
		r.With(middleware.RateLimit(limiter, middleware.RateKey("auth_forgot"), 5, time.Hour)).
			Post("/forgot-password", h.ForgotPassword)
		r.With(middleware.RateLimit(limiter, middleware.RateKey("auth_reset"), 10, time.Hour)).
			Post("/reset-password", h.ResetPassword)

		r.Group(func(r chi.Router) {
			r.Use(Authenticate(signer))
			r.Get("/me", h.Me)
			r.Post("/logout-all", h.LogoutAll)
		})
	})
}
