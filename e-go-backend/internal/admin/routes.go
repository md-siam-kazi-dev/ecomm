package admin

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/internal/audit"
)

// Register mounts the admin routes. The caller is responsible for wrapping
// this router with authentication + RequireRole("admin").
func Register(r chi.Router, db *pgxpool.Pool) {
	repo := NewRepository(db)

	h := NewHandler(NewService(repo))
	r.Get("/users", h.ListUsers)

	products := NewProductHandler(NewProductService(repo))
	r.Get("/products", products.ListProducts)
	r.Get("/products/{id}", products.GetProduct)
	r.Post("/products/{id}/approve", products.Approve)
	r.Post("/products/{id}/reject", products.Reject)
	r.Post("/products/{id}/changes", products.RequestChanges)

	// Admin audit trail (read-only).
	audit.Register(r, db)
}
