package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/md-siam-kazi/e-go-backend/internal/config"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
)

// Server owns the HTTP router and its dependencies. It is intentionally
// transport-only: business logic lives in the domain service layers.
type Server struct {
	cfg    *config.Config
	log    *slog.Logger
	db     *pgxpool.Pool
	redis  *redis.Client
	signer *jwt.Signer
	router chi.Router
}

func New(cfg *config.Config, log *slog.Logger, db *pgxpool.Pool, rdb *redis.Client) *Server {
	s := &Server{
		cfg:   cfg,
		log:   log,
		db:    db,
		redis: rdb,
		// One signer shared by every module that authenticates requests.
		signer: jwt.NewSigner(
			cfg.JWT.AccessSecret,
			cfg.JWT.Issuer,
			cfg.JWT.Audience,
			cfg.JWT.AccessTTL,
		),
	}
	s.router = s.buildRouter()
	return s
}

// Handler exposes the underlying router for use with httptest.
func (s *Server) Handler() http.Handler {
	return s.router
}

// HTTPServer builds the *http.Server with the configured timeouts. The caller
// is responsible for starting and gracefully shutting it down.
func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:              ":" + s.cfg.App.Port,
		Handler:           s.router,
		ReadTimeout:       s.cfg.App.ReadTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      s.cfg.App.WriteTimeout,
		IdleTimeout:       s.cfg.App.IdleTimeout,
	}
}
