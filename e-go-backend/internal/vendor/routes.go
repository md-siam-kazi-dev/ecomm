package vendor

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register mounts the vendor routes. The caller wraps this router with
// authentication + RequireRole("vendor").
func Register(r chi.Router, db *pgxpool.Pool, uploader ImageUploader) {
	mount(r, NewHandler(NewService(NewRepository(db), uploader)))
}

// mount wires the handlers. Split out so tests can inject fakes.
func mount(r chi.Router, h *Handler) {
	r.Get("/profile", h.Profile)
	r.Get("/dashboard", h.Dashboard)
	r.Get("/products", h.ListProducts)
	r.Post("/products", h.CreateProduct)
	r.Post("/products/images", h.UploadImage)
	r.Get("/products/{id}", h.GetProduct)
	r.Put("/products/{id}", h.UpdateProduct)
	r.Get("/orders", h.ListOrders)
	r.Get("/orders/{id}", h.GetOrder)
	r.Post("/orders/{id}/shipment", h.AddShipmentEvent)
	r.Get("/revenue", h.Revenue)
	r.Get("/inventory", h.ListInventory)
	r.Post("/inventory/{variantId}/adjust", h.AdjustStock)
	r.Get("/reviews", h.ListReviews)
	r.Get("/reviews/summary", h.ReviewSummary)
	r.Post("/reviews/{id}/reply", h.ReplyToReview)
	r.Get("/questions", h.ListQuestions)
	r.Post("/questions/{id}/answer", h.AnswerQuestion)
}
