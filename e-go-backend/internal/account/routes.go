package account

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register mounts the account routes. The caller wraps this router with
// authentication.
func Register(r chi.Router, db *pgxpool.Pool, uploader ImageUploader) {
	mount(r, NewHandler(NewService(NewRepository(db), uploader)))
}

// mount wires the handlers. Split out so tests can inject fakes.
func mount(r chi.Router, h *Handler) {
	r.Get("/profile", h.GetProfile)
	r.Put("/profile", h.UpdateProfile)
	r.Post("/avatar", h.UploadAvatar)

	r.Route("/addresses", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Post("/{id}/default", h.SetDefault)
	})
}
