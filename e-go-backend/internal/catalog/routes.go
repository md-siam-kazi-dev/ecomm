package catalog

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register mounts the public catalog routes. Mount at the top level of
// /api/v1 with NO auth middleware — every route is public (read-only).
func Register(r chi.Router, db *pgxpool.Pool) {
	mount(r, NewHandler(NewService(NewRepository(db))))
}

// mount wires the handlers. Split out so tests can inject fakes.
func mount(r chi.Router, h *Handler) {
	r.Get("/products", h.ListProducts)
	r.Get("/products/{slug}", h.GetProduct)
	r.Get("/products/{slug}/reviews", h.ListReviews)
	r.Get("/products/{slug}/rating", h.Rating)
	r.Get("/products/{slug}/questions", h.Questions)
	r.Get("/products/{slug}/related", h.Related)
	r.Get("/stores/{slug}", h.GetStore)
	r.Get("/stores/{slug}/products", h.StoreProducts)
	r.Get("/categories", h.ListCategories)
	r.Get("/reviews/top", h.TopReviews)
	// Public, read-only keyset scans for the frontend sitemap (no auth).
	r.Get("/sitemap/products", h.SitemapProducts)
	r.Get("/sitemap/stores", h.SitemapStores)
}
