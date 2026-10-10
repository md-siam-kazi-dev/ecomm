package order

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/payment"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// maxWebhookBodyBytes bounds a provider callback body.
const maxWebhookBodyBytes = 1 << 20

// Handler translates order HTTP requests into service calls.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// CreateOrder handles POST /api/v1/orders.
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	out, err := h.svc.Create(r.Context(), id.UserID, req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, out, "Order placed")
}

// ListOrders handles GET /api/v1/orders.
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	items, pagination, err := h.svc.List(r.Context(), id.UserID, page, limit)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// GetOrder handles GET /api/v1/orders/{id}.
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	o, err := h.svc.Get(r.Context(), id.UserID, chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, o, "")
}

// Webhook handles POST /api/v1/orders/webhook/{provider}. Public, but only a
// payload the provider verifies can confirm an order.
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBodyBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	verify := payment.VerifyRequest{
		Signature: r.Header.Get("Stripe-Signature"),
		RawBody:   raw,
	}
	if err := h.svc.HandleWebhook(r.Context(), chi.URLParam(r, "provider"), verify); err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, map[string]string{"received": "true"}, "")
}
