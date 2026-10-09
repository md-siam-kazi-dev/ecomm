package storeapp

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/internal/config"
	"github.com/md-siam-kazi/e-go-backend/pkg/imgbb"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
	"github.com/md-siam-kazi/e-go-backend/pkg/middleware"
	"github.com/md-siam-kazi/e-go-backend/pkg/payment"
	"github.com/md-siam-kazi/e-go-backend/pkg/ratelimit"
)

// BuildService constructs the module service from configuration.
func BuildService(db *pgxpool.Pool, cfg *config.Config, registry *payment.Registry) *Service {
	svc := NewService(NewRepository(db), registry, Config{
		FeeEnabled:   cfg.Payment.ApplicationFeeEnabled,
		FeeAmount:    cfg.Payment.ApplicationFeeAmount,
		FeeCurrency:  cfg.Payment.ApplicationFeeCurrency,
		ProviderName: cfg.Payment.Provider,
		PlatformName: cfg.App.Name,
	})
	svc.uploader = imgbb.New(cfg.Images.ImgBBAPIKey)
	return svc
}

// RegisterApplicant mounts the applicant + payment routes (auth required) and
// the public provider webhook.
func RegisterApplicant(api chi.Router, signer *jwt.Signer, svc *Service, limiter ratelimit.Limiter) {
	h := NewHandler(svc)

	// Public: verified server-to-server provider callback.
	api.Post("/payments/webhook/{provider}", h.Webhook)

	api.Group(func(r chi.Router) {
		r.Use(auth.Authenticate(signer))

		r.With(middleware.RateLimit(limiter, middleware.RateKey("store_application"), 5, time.Hour)).
			Post("/store-applications", h.Submit)
		r.With(middleware.RateLimit(limiter, middleware.RateKey("store_image"), 30, time.Hour)).
			Post("/store-applications/images", h.UploadImage)
		r.Get("/store-applications/me", h.MyApplications)
		r.Get("/store-applications/{id}", h.Detail)
		r.Get("/store-applications/{id}/approval-letter", h.ApprovalLetter)
		r.With(middleware.RateLimit(limiter, middleware.RateKey("store_payment"), 10, time.Hour)).
			Post("/store-applications/{id}/payments", h.InitiatePayment)
		r.Get("/store-applications/{id}/payments", h.Payments)
		r.Get("/my-stores", h.MyStores)
	})
}

// RegisterAdmin mounts the admin routes. The caller wraps this router with
// auth.Authenticate + RequireRole("admin").
func RegisterAdmin(r chi.Router, svc *Service) {
	h := NewAdminHandler(svc)

	r.Get("/store-applications", h.ListApplications)
	r.Get("/store-applications/{id}", h.GetApplication)
	r.Post("/store-applications/{id}/approve", h.Approve)
	r.Post("/store-applications/{id}/reject", h.Reject)
	r.Post("/store-applications/{id}/notes", h.UpdateNotes)
	r.Get("/stores", h.ListStores)
	r.Get("/application-payments", h.ListPayments)
}
