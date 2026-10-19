package follow

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
)

// Register mounts the authenticated store-follow routes (any role).
func Register(api chi.Router, signer *jwt.Signer, db *pgxpool.Pool) {
	mount(api, NewHandler(NewService(NewRepository(db))), signer)
}

// mount wires the handlers. Split out so tests can inject fakes.
func mount(api chi.Router, h *Handler, signer *jwt.Signer) {
	// Reading the follow state (and the public follower count) works for guests.
	api.With(auth.OptionalAuthenticate(signer)).
		Get("/stores/{slug}/follow", h.Status)

	// Following/unfollowing requires an authenticated user.
	api.Group(func(r chi.Router) {
		r.Use(auth.Authenticate(signer))
		r.Post("/stores/{slug}/follow", h.Follow)
		r.Delete("/stores/{slug}/follow", h.Unfollow)
	})
}
