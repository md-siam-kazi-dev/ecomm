package vendor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/audit"
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
	// audit is optional; a nil recorder (e.g. in unit tests) disables auditing.
	audit *audit.Recorder
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// record writes a best-effort audit event for a successful vendor action.
func (h *Handler) record(ctx context.Context, e audit.Event) {
	if h.audit == nil {
		return
	}
	_ = h.audit.Log(ctx, e)
}

// vendorEvent builds an event attributed to the calling vendor.
func vendorEvent(userID, action, entityType, entityID string, newValues map[string]any) audit.Event {
	return audit.Event{
		ActorID:    &userID,
		ActorRole:  audit.RoleVendor,
		Action:     action,
		EntityType: entityType,
		EntityID:   &entityID,
		Result:     audit.ResultSuccess,
		NewValues:  newValues,
	}
}

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
	h.record(r.Context(), vendorEvent(id.UserID, audit.ActionVendorProductCreated, audit.EntityProduct, product.ID, map[string]any{
		"name": product.Name, "sku": product.SKU, "status": product.Status,
	}))
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
	h.record(r.Context(), vendorEvent(id.UserID, audit.ActionVendorProductUpdated, audit.EntityProduct, product.ID, map[string]any{
		"name": product.Name, "sku": product.SKU, "status": product.Status,
	}))
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
	h.record(r.Context(), audit.Event{
		ActorID:    &id.UserID,
		ActorRole:  audit.RoleVendor,
		Action:     audit.ActionVendorInventoryAdjusted,
		EntityType: audit.EntityVariant,
		EntityID:   &item.VariantID,
		Result:     audit.ResultSuccess,
		Reason:     req.Note,
		NewValues:  map[string]any{"quantity": item.Quantity, "product_id": item.ProductID},
	})
	response.OK(w, item, "Stock updated")
}

// ListInventoryMovements handles GET /api/v1/vendor/inventory/movements.
func (h *Handler) ListInventoryMovements(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	items, pagination, err := h.svc.ListInventoryMovements(r.Context(), id.UserID, requestedStoreID(r), queryFrom(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// SetInventoryThreshold handles PUT /api/v1/vendor/inventory/{variantId}/threshold.
func (h *Handler) SetInventoryThreshold(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req InventoryThresholdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	item, err := h.svc.SetInventoryThreshold(r.Context(), id.UserID, requestedStoreID(r), chi.URLParam(r, "variantId"), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, item, "Threshold updated")
}

// ExportOrders handles GET /api/v1/vendor/orders/export.
func (h *Handler) ExportOrders(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}
	format, ok := exportFormat(w, r)
	if !ok {
		return
	}
	q, err := exportQuery(r)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	rows, err := h.svc.ExportOrders(r.Context(), id.UserID, requestedStoreID(r), q)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	_ = writeExport(w, format, ordersExportFile(rows))
}

// ExportInventory handles GET /api/v1/vendor/inventory/export.
func (h *Handler) ExportInventory(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}
	format, ok := exportFormat(w, r)
	if !ok {
		return
	}

	rows, err := h.svc.ExportInventory(r.Context(), id.UserID, requestedStoreID(r), queryFrom(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	_ = writeExport(w, format, inventoryExportFile(rows))
}

// errInvalidExportFormat is returned for an unsupported export format.
var errInvalidExportFormat = apperr.New(http.StatusBadRequest, "INVALID_FORMAT", "Format must be csv or xlsx")

// exportFormat resolves the requested export format, defaulting to csv.
func exportFormat(w http.ResponseWriter, r *http.Request) (string, bool) {
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "xlsx" {
		response.Error(w, r, errInvalidExportFormat)
		return "", false
	}
	return format, true
}

// exportQuery reads the optional status and inclusive from/to date filters.
func exportQuery(r *http.Request) (ExportQuery, error) {
	q := r.URL.Query()
	var out ExportQuery
	out.Status = q.Get("status")
	if v := strings.TrimSpace(q.Get("from")); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return ExportQuery{}, apperr.ErrValidation
		}
		out.From = t
	}
	if v := strings.TrimSpace(q.Get("to")); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return ExportQuery{}, apperr.ErrValidation
		}
		// `to` is inclusive: match anything before the next day.
		out.To = t.AddDate(0, 0, 1)
	}
	return out, nil
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
	h.record(r.Context(), vendorEvent(id.UserID, audit.ActionVendorShipmentUpdated, audit.EntityOrder, order.ID, map[string]any{
		"status": req.Status, "carrier": req.Carrier, "tracking_number": req.TrackingNumber,
	}))
	response.OK(w, order, "Shipment updated")
}

// SetOrderStatus handles POST /api/v1/vendor/orders/{id}/status.
func (h *Handler) SetOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req OrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	order, err := h.svc.SetOrderStatus(r.Context(), id.UserID, requestedStoreID(r), chi.URLParam(r, "id"), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	h.record(r.Context(), vendorEvent(id.UserID, audit.ActionVendorOrderStatusChanged, audit.EntityOrder, order.ID, map[string]any{
		"status": order.Status,
	}))
	response.OK(w, order, "Order updated")
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
	h.record(r.Context(), vendorEvent(id.UserID, audit.ActionVendorReviewReplied, audit.EntityReview, review.ID, nil))
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
	h.record(r.Context(), vendorEvent(id.UserID, audit.ActionVendorQuestionAnswered, audit.EntityQuestion, question.ID, nil))
	response.OK(w, question, "Answer saved")
}

func queryFrom(r *http.Request) ListQuery {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	return ListQuery{
		Page:      page,
		Limit:     limit,
		Search:    q.Get("search"),
		Status:    q.Get("status"),
		VariantID: q.Get("variant_id"),
		Type:      q.Get("type"),
	}
}
