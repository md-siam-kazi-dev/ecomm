package settings

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterPublic mounts the read-only storefront settings endpoint. It has NO
// auth middleware: the announcement bar renders for anonymous visitors.
func RegisterPublic(r chi.Router, db *pgxpool.Pool) {
	mountPublic(r, NewHandler(NewService(NewRepository(db))))
}

// RegisterAdmin mounts the admin settings endpoints. The caller (the /admin
// group in server/routes.go) is responsible for wrapping them with
// authentication + RequireRole("admin").
func RegisterAdmin(r chi.Router, db *pgxpool.Pool) {
	mountAdmin(r, NewHandler(NewService(NewRepository(db))))
}

// mountPublic/mountAdmin are split out so tests can inject fakes.
func mountPublic(r chi.Router, h *Handler) {
	r.Get("/storefront/settings", h.Get)
}

func mountAdmin(r chi.Router, h *Handler) {
	r.Route("/settings", func(r chi.Router) {
		r.Get("/", h.Get)
		r.Put("/", h.Update)
	})
}
