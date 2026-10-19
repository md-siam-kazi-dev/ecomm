package audit

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register mounts the admin audit routes. The caller (admin.Register) is
// responsible for wrapping them with authentication + RequireRole("admin"), so
// the endpoints are admin-only by construction.
func Register(r chi.Router, db *pgxpool.Pool) {
	h := NewHandler(NewService(NewRepository(db)))

	r.Route("/audit-logs", func(r chi.Router) {
		r.Get("/", h.List)
		r.Get("/summary", h.Summary)
		r.Get("/{id}", h.Get)
	})
}
