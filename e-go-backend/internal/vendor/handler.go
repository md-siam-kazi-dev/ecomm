package vendor

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

var errInvalidBody = apperr.New(http.StatusBadRequest, "BAD_REQUEST", "Request body must be valid JSON")

// isAllowedImage sniffs the bytes so only real images are uploaded.
func isAllowedImage(data []byte) bool {
	switch http.DetectContentType(data) {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

// Handler translates vendor HTTP requests into service calls. The authenticated
// identity is expected in the request context (auth.Authenticate).
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func identity(r *http.Request) (auth.Identity, bool) {
	return auth.IdentityFromContext(r.Context())
}

// requestedStoreID is the optional store the vendor is operating. It lets one
// account manage several stores; when omitted the service picks the caller's
// active store.
func requestedStoreID(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Store-Id")); v != "" {
		return v
	}
	return strings.TrimSpace(r.URL.Query().Get("store_id"))
}

// Profile handles GET /api/v1/vendor/profile.
func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	profile, err := h.svc.Profile(r.Context(), id.UserID, requestedStoreID(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, profile, "")
}

// Dashboard handles GET /api/v1/vendor/dashboard.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	summary, err := h.svc.Dashboard(r.Context(), id.UserID, requestedStoreID(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, summary, "")
}

// ListProducts handles GET /api/v1/vendor/products.
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	items, pagination, err := h.svc.ListProducts(r.Context(), id.UserID, requestedStoreID(r), queryFrom(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// GetProduct handles GET /api/v1/vendor/products/{id}.
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	product, err := h.svc.GetProduct(r.Context(), id.UserID, requestedStoreID(r), chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, product, "")
}

// CreateProduct handles POST /api/v1/vendor/products.
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	product, err := h.svc.CreateProduct(r.Context(), id.UserID, requestedStoreID(r), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, product, "Product submitted")
}

// UpdateProduct handles PUT /api/v1/vendor/products/{id}.
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	product, err := h.svc.UpdateProduct(r.Context(), id.UserID, requestedStoreID(r), chi.URLParam(r, "id"), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, product, "Product updated")
}

// UploadImage handles POST /api/v1/vendor/products/images (multipart, "image").
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if _, ok := identity(r); !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxProductImageBytes+1024)
	if err := r.ParseMultipartForm(maxProductImageBytes); err != nil {
		response.Error(w, r, errInvalidImage)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		response.Error(w, r, errInvalidImage)
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, maxProductImageBytes+1))
	if err != nil {
		response.Error(w, r, apperr.ErrInternal.WithCause(err))
		return
	}
	if !isAllowedImage(data) {
		response.Error(w, r, errInvalidImage)
		return
	}

	url, err := h.svc.UploadImage(r.Context(), header.Filename, data)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, ImageUploadResponse{URL: url}, "Image uploaded")
}

// ListOrders handles GET /api/v1/vendor/orders.
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	items, pagination, err := h.svc.ListOrders(r.Context(), id.UserID, requestedStoreID(r), queryFrom(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// ListInventory handles GET /api/v1/vendor/inventory.
func (h *Handler) ListInventory(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	items, pagination, err := h.svc.ListInventory(r.Context(), id.UserID, requestedStoreID(r), queryFrom(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// AdjustStock handles POST /api/v1/vendor/inventory/{variantId}/adjust.
func (h *Handler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req InventoryAdjustRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	item, err := h.svc.AdjustStock(r.Context(), id.UserID, requestedStoreID(r), chi.URLParam(r, "variantId"), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, item, "Stock updated")
}

// GetOrder handles GET /api/v1/vendor/orders/{id}.
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	order, err := h.svc.OrderDetail(r.Context(), id.UserID, requestedStoreID(r), chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, order, "")
}

// AddShipmentEvent handles POST /api/v1/vendor/orders/{id}/shipment.
func (h *Handler) AddShipmentEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req ShipmentEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	order, err := h.svc.AddShipmentEvent(r.Context(), id.UserID, requestedStoreID(r), chi.URLParam(r, "id"), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, order, "Shipment updated")
}

// Revenue handles GET /api/v1/vendor/revenue.
func (h *Handler) Revenue(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	revenue, err := h.svc.Revenue(r.Context(), id.UserID, requestedStoreID(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, revenue, "")
}

// ListReviews handles GET /api/v1/vendor/reviews.
func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	items, pagination, err := h.svc.ListReviews(r.Context(), id.UserID, requestedStoreID(r), queryFrom(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// ReviewSummary handles GET /api/v1/vendor/reviews/summary.
func (h *Handler) ReviewSummary(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	summary, err := h.svc.ReviewSummary(r.Context(), id.UserID, requestedStoreID(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, summary, "")
}

// ReplyToReview handles POST /api/v1/vendor/reviews/{id}/reply.
func (h *Handler) ReplyToReview(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req ReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	review, err := h.svc.ReplyToReview(r.Context(), id.UserID, requestedStoreID(r), chi.URLParam(r, "id"), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, review, "Reply saved")
}

// ListQuestions handles GET /api/v1/vendor/questions.
func (h *Handler) ListQuestions(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	items, pagination, err := h.svc.ListQuestions(r.Context(), id.UserID, requestedStoreID(r), queryFrom(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// AnswerQuestion handles POST /api/v1/vendor/questions/{id}/answer.
func (h *Handler) AnswerQuestion(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req AnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	question, err := h.svc.AnswerQuestion(r.Context(), id.UserID, requestedStoreID(r), chi.URLParam(r, "id"), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, question, "Answer saved")
}

func queryFrom(r *http.Request) ListQuery {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	return ListQuery{
		Page:   page,
		Limit:  limit,
		Search: q.Get("search"),
		Status: q.Get("status"),
	}
}
