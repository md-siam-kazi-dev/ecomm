package vendor

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/imgbb"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

const (
	defaultLimit = 20
	maxSearchLen = 100
	// maxProductImageBytes caps an uploaded product image (5 MB).
	maxProductImageBytes = 5 << 20
	// maxProductImages caps how many images a product may have.
	maxProductImages = 8
	// maxNoteLength bounds an inventory adjustment note.
	maxNoteLength = 500
)

var (
	errInvalidImage = apperr.New(http.StatusBadRequest, "INVALID_IMAGE", "Attach a PNG, JPG, WebP or GIF image (max 5 MB)")
	// errImageUnavailable is returned when image hosting is not configured.
	errImageUnavailable = apperr.New(http.StatusServiceUnavailable, "IMAGE_HOSTING_UNAVAILABLE", "Image uploads are not configured")
)

// allowedLimits mirrors the admin list policy.
var allowedLimits = map[int]struct{}{10: {}, 20: {}, 50: {}, 100: {}}

// allowedOrderStatuses are the valid values for the vendor order status filter.
var allowedOrderStatuses = map[string]struct{}{
	"pending": {}, "confirmed": {}, "processing": {}, "packed": {},
	"shipped": {}, "delivered": {}, "cancelled": {}, "returned": {}, "refunded": {},
}

// allowedStockFilters are the valid values for the inventory stock filter.
var allowedStockFilters = map[string]struct{}{"in": {}, "low": {}, "out": {}}

// errInvalidStock is returned for a bad stock adjustment.
var errInvalidStock = apperr.New(http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Enter a whole stock count of 0 or more")

// allowedShipmentStatuses are the valid values for a shipment/tracking update.
var allowedShipmentStatuses = map[string]struct{}{
	"pending": {}, "packed": {}, "shipped": {}, "in_transit": {},
	"out_for_delivery": {}, "delivered": {}, "failed": {}, "returned": {},
}

const (
	maxCarrierLen     = 150
	maxTrackingLen    = 255
	maxLocationLen    = 255
	maxDescriptionLen = 500
)

// errInvalidShipment is returned for a bad shipment/tracking update.
var errInvalidShipment = apperr.New(http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Choose a valid shipment status")

// ImageUploader hosts an image and returns its public URL.
type ImageUploader interface {
	Upload(ctx context.Context, filename string, data []byte) (string, error)
}

// Service owns the vendor business logic (Handler → Service → Repository).
type Service struct {
	repo     VendorRepository
	uploader ImageUploader
}

func NewService(repo VendorRepository, uploader ImageUploader) *Service {
	return &Service{repo: repo, uploader: uploader}
}

// scope resolves the store the caller may operate. Vendor requests are
// store-scoped: the user must be an active member of an active store.
func (s *Service) scope(ctx context.Context, userID, storeID string) (StoreScope, error) {
	return s.repo.ResolveStore(ctx, userID, storeID)
}

// Profile returns the store the caller is operating.
func (s *Service) Profile(ctx context.Context, userID, storeID string) (ProfileResponse, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return ProfileResponse{}, err
	}
	return ProfileResponse{
		ID:           st.VendorID,
		StoreID:      st.StoreID,
		BusinessName: st.Name,
		Status:       st.Status,
		Role:         st.Role,
	}, nil
}

// Dashboard returns the store's KPIs.
func (s *Service) Dashboard(ctx context.Context, userID, storeID string) (SummaryResponse, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return SummaryResponse{}, err
	}
	summary, err := s.repo.Dashboard(ctx, st.VendorID)
	if err != nil {
		return SummaryResponse{}, err
	}
	return toSummaryResponse(summary), nil
}

// ListProducts returns a page of the store's products.
func (s *Service) ListProducts(ctx context.Context, userID, storeID string, q ListQuery) ([]ProductItem, response.Pagination, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return nil, response.Pagination{}, err
	}

	page, limit := normalizePagination(q)
	records, total, err := s.repo.ListProducts(ctx, st.VendorID, ListQuery{
		Page: page, Limit: limit, Search: normalizeSearch(q.Search),
	})
	if err != nil {
		return nil, response.Pagination{}, err
	}

	items := make([]ProductItem, 0, len(records))
	for _, rec := range records {
		items = append(items, toProductItem(rec))
	}
	return items, response.Pagination{Page: page, Limit: limit, Total: total}, nil
}

// GetProduct returns one of the store's products.
func (s *Service) GetProduct(ctx context.Context, userID, storeID, productID string) (ProductDetailResponse, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return ProductDetailResponse{}, err
	}
	detail, err := s.repo.GetProduct(ctx, st.VendorID, productID)
	if err != nil {
		return ProductDetailResponse{}, err
	}
	return toProductDetailResponse(detail), nil
}

// CreateProduct validates and stores a new product for the store.
func (s *Service) CreateProduct(ctx context.Context, userID, storeID string, req ProductRequest) (ProductDetailResponse, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return ProductDetailResponse{}, err
	}
	params, err := validateProduct(req)
	if err != nil {
		return ProductDetailResponse{}, err
	}
	detail, err := s.repo.CreateProduct(ctx, st.VendorID, params)
	if err != nil {
		return ProductDetailResponse{}, err
	}
	return toProductDetailResponse(detail), nil
}

// UpdateProduct validates and edits one of the store's products. The product
// returns to pending review.
func (s *Service) UpdateProduct(ctx context.Context, userID, storeID, productID string, req ProductRequest) (ProductDetailResponse, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return ProductDetailResponse{}, err
	}
	params, err := validateProduct(req)
	if err != nil {
		return ProductDetailResponse{}, err
	}
	detail, err := s.repo.UpdateProduct(ctx, st.VendorID, productID, params)
	if err != nil {
		return ProductDetailResponse{}, err
	}
	return toProductDetailResponse(detail), nil
}

// UploadImage hosts a product image and returns its URL.
func (s *Service) UploadImage(ctx context.Context, filename string, data []byte) (string, error) {
	if len(data) == 0 || len(data) > maxProductImageBytes {
		return "", errInvalidImage
	}
	url, err := s.uploader.Upload(ctx, filename, data)
	if err != nil {
		if errors.Is(err, imgbb.ErrNotConfigured) {
			return "", errImageUnavailable
		}
		return "", apperr.ErrInternal.WithCause(err)
	}
	return url, nil
}

// ListOrders returns a page of orders containing the store's items.
func (s *Service) ListOrders(ctx context.Context, userID, storeID string, q ListQuery) ([]OrderItem, response.Pagination, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return nil, response.Pagination{}, err
	}

	status := strings.ToLower(strings.TrimSpace(q.Status))
	if _, ok := allowedOrderStatuses[status]; !ok {
		status = ""
	}

	page, limit := normalizePagination(q)
	records, total, err := s.repo.ListOrders(ctx, st.VendorID, ListQuery{Page: page, Limit: limit, Status: status})
	if err != nil {
		return nil, response.Pagination{}, err
	}

	items := make([]OrderItem, 0, len(records))
	for _, rec := range records {
		items = append(items, toOrderItem(rec))
	}
	return items, response.Pagination{Page: page, Limit: limit, Total: total}, nil
}

// ListInventory returns a page of the store's inventory (per variant).
func (s *Service) ListInventory(ctx context.Context, userID, storeID string, q ListQuery) ([]InventoryItem, response.Pagination, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return nil, response.Pagination{}, err
	}

	page, limit := normalizePagination(q)
	records, total, err := s.repo.ListInventory(ctx, st.VendorID, ListQuery{
		Page: page, Limit: limit, Search: normalizeSearch(q.Search), Status: normalizeStockFilter(q.Status),
	})
	if err != nil {
		return nil, response.Pagination{}, err
	}

	items := make([]InventoryItem, 0, len(records))
	for _, rec := range records {
		items = append(items, toInventoryItem(rec))
	}
	return items, response.Pagination{Page: page, Limit: limit, Total: total}, nil
}

// AdjustStock sets the on-hand stock for a variant the store owns.
func (s *Service) AdjustStock(ctx context.Context, userID, storeID, variantID string, req InventoryAdjustRequest) (InventoryItem, error) {
	if req.Quantity < 0 {
		return InventoryItem{}, errInvalidStock
	}

	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return InventoryItem{}, err
	}

	note := strings.TrimSpace(req.Note)
	if len(note) > maxNoteLength {
		note = note[:maxNoteLength]
	}

	rec, err := s.repo.AdjustStock(ctx, st.VendorID, variantID, userID, req.Quantity, note)
	if err != nil {
		return InventoryItem{}, err
	}
	return toInventoryItem(rec), nil
}

// OrderDetail returns a store order with its items, address and shipment timeline.
func (s *Service) OrderDetail(ctx context.Context, userID, storeID, orderID string) (OrderDetailResponse, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return OrderDetailResponse{}, err
	}
	rec, err := s.repo.OrderDetail(ctx, st.VendorID, orderID)
	if err != nil {
		return OrderDetailResponse{}, err
	}
	return toOrderDetailResponse(rec), nil
}

// AddShipmentEvent records a tracking update for the store's shipment and
// returns the refreshed order (timeline + advanced status).
func (s *Service) AddShipmentEvent(ctx context.Context, userID, storeID, orderID string, req ShipmentEventRequest) (OrderDetailResponse, error) {
	params, err := validateShipment(req)
	if err != nil {
		return OrderDetailResponse{}, err
	}
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return OrderDetailResponse{}, err
	}
	if err := s.repo.UpsertShipmentEvent(ctx, st.VendorID, orderID, params); err != nil {
		return OrderDetailResponse{}, err
	}
	rec, err := s.repo.OrderDetail(ctx, st.VendorID, orderID)
	if err != nil {
		return OrderDetailResponse{}, err
	}
	return toOrderDetailResponse(rec), nil
}

// Revenue returns the store's revenue aggregate.
func (s *Service) Revenue(ctx context.Context, userID, storeID string) (RevenueResponse, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return RevenueResponse{}, err
	}
	summary, err := s.repo.Revenue(ctx, st.VendorID)
	if err != nil {
		return RevenueResponse{}, err
	}
	return toRevenueResponse(summary), nil
}

// ---- reviews --------------------------------------------------------------

// allowedReviewFilters are the valid values for the vendor review filter.
var allowedReviewFilters = map[string]struct{}{"all": {}, "replied": {}, "unreplied": {}}

// allowedQuestionFilters are the valid values for the vendor question filter.
var allowedQuestionFilters = map[string]struct{}{"all": {}, "unanswered": {}, "answered": {}}

// maxReplyLen bounds a vendor review reply.
const maxReplyLen = 1000

// maxAnswerLen bounds a vendor question answer.
const maxAnswerLen = 1000

// ListReviews returns a page of reviews on the store's products.
func (s *Service) ListReviews(ctx context.Context, userID, storeID string, q ListQuery) ([]VendorReviewResponse, response.Pagination, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return nil, response.Pagination{}, err
	}

	page, limit := normalizePagination(q)
	records, total, err := s.repo.ListReviews(ctx, st.VendorID, ListQuery{
		Page: page, Limit: limit, Search: normalizeSearch(q.Search), Status: normalizeReviewFilter(q.Status),
	})
	if err != nil {
		return nil, response.Pagination{}, err
	}

	items := make([]VendorReviewResponse, 0, len(records))
	for _, rec := range records {
		items = append(items, toVendorReviewResponse(rec))
	}
	return items, response.Pagination{Page: page, Limit: limit, Total: total}, nil
}

// ReviewSummary returns the store's aggregate star feedback.
func (s *Service) ReviewSummary(ctx context.Context, userID, storeID string) (ReviewSummaryResponse, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return ReviewSummaryResponse{}, err
	}
	rec, err := s.repo.ReviewSummary(ctx, st.VendorID)
	if err != nil {
		return ReviewSummaryResponse{}, err
	}
	return toReviewSummaryResponse(rec), nil
}

// ReplyToReview sets or replaces the store's reply on one of its reviews.
func (s *Service) ReplyToReview(ctx context.Context, userID, storeID, reviewID string, req ReplyRequest) (VendorReviewResponse, error) {
	reply := strings.TrimSpace(req.Reply)
	if reply == "" || len(reply) > maxReplyLen {
		return VendorReviewResponse{}, errInvalidReply
	}

	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return VendorReviewResponse{}, err
	}
	rec, err := s.repo.ReplyToReview(ctx, st.VendorID, reviewID, reply)
	if err != nil {
		return VendorReviewResponse{}, err
	}
	return toVendorReviewResponse(rec), nil
}

// ---- questions ------------------------------------------------------------

// ListQuestions returns a page of questions on the store's products.
func (s *Service) ListQuestions(ctx context.Context, userID, storeID string, q ListQuery) ([]VendorQuestionResponse, response.Pagination, error) {
	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return nil, response.Pagination{}, err
	}

	page, limit := normalizePagination(q)
	records, total, err := s.repo.ListQuestions(ctx, st.VendorID, ListQuery{
		Page: page, Limit: limit, Search: normalizeSearch(q.Search), Status: normalizeQuestionFilter(q.Status),
	})
	if err != nil {
		return nil, response.Pagination{}, err
	}

	items := make([]VendorQuestionResponse, 0, len(records))
	for _, rec := range records {
		items = append(items, toVendorQuestionResponse(rec))
	}
	return items, response.Pagination{Page: page, Limit: limit, Total: total}, nil
}

// AnswerQuestion upserts the store's official answer for one of its questions.
func (s *Service) AnswerQuestion(ctx context.Context, userID, storeID, questionID string, req AnswerRequest) (VendorQuestionResponse, error) {
	answer := strings.TrimSpace(req.Answer)
	if answer == "" || len(answer) > maxAnswerLen {
		return VendorQuestionResponse{}, errInvalidAnswer
	}

	st, err := s.scope(ctx, userID, storeID)
	if err != nil {
		return VendorQuestionResponse{}, err
	}
	rec, err := s.repo.AnswerQuestion(ctx, st.VendorID, questionID, userID, answer)
	if err != nil {
		return VendorQuestionResponse{}, err
	}
	return toVendorQuestionResponse(rec), nil
}

// ---- validation -----------------------------------------------------------

func validateShipment(req ShipmentEventRequest) (ShipmentParams, error) {
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if _, ok := allowedShipmentStatuses[status]; !ok {
		return ShipmentParams{}, errInvalidShipment
	}

	params := ShipmentParams{
		Carrier:        trimBound(req.Carrier, maxCarrierLen),
		TrackingNumber: trimBound(req.TrackingNumber, maxTrackingLen),
		Status:         status,
		Location:       trimBound(req.Location, maxLocationLen),
		Description:    trimBound(req.Description, maxDescriptionLen),
	}
	if raw := strings.TrimSpace(req.EventTime); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return ShipmentParams{}, apperr.ErrValidation
		}
		params.EventTime = &parsed
	}
	return params, nil
}

func trimBound(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) > max {
		s = s[:max]
	}
	return s
}

func normalizeStockFilter(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if _, ok := allowedStockFilters[v]; !ok {
		return ""
	}
	return v
}

func normalizeReviewFilter(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if _, ok := allowedReviewFilters[v]; !ok {
		return ""
	}
	return v
}

func normalizeQuestionFilter(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if _, ok := allowedQuestionFilters[v]; !ok {
		return ""
	}
	return v
}

func validateProduct(req ProductRequest) (ProductParams, error) {
	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)
	category := strings.TrimSpace(req.Category)

	if len(name) < 2 || len(name) > 255 {
		return ProductParams{}, apperr.ErrValidation
	}
	if description == "" {
		return ProductParams{}, apperr.ErrValidation
	}
	if category == "" {
		return ProductParams{}, apperr.ErrValidation
	}
	if req.Price <= 0 {
		return ProductParams{}, apperr.ErrValidation
	}
	if req.SalePrice != nil && (*req.SalePrice <= 0 || *req.SalePrice >= req.Price) {
		return ProductParams{}, apperr.ErrValidation
	}
	if req.Stock < 0 || req.LowStockThreshold < 0 {
		return ProductParams{}, apperr.ErrValidation
	}
	if len(req.Images) == 0 || len(req.Images) > maxProductImages {
		return ProductParams{}, apperr.ErrValidation
	}

	attrs := map[string]string{}
	for key, value := range req.Attributes {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			attrs[key] = trimmed
		}
	}

	return ProductParams{
		Name:              name,
		Description:       description,
		ShortDescription:  strings.TrimSpace(req.ShortDescription),
		CategorySlug:      category,
		Brand:             strings.TrimSpace(req.Brand),
		SKU:               strings.TrimSpace(req.SKU),
		Price:             req.Price,
		SalePrice:         req.SalePrice,
		Stock:             req.Stock,
		LowStockThreshold: req.LowStockThreshold,
		Images:            req.Images,
		Attributes:        attrs,
		Submit:            req.Submit,
	}, nil
}

func normalizePagination(q ListQuery) (page, limit int) {
	page = q.Page
	if page < 1 {
		page = 1
	}
	limit = q.Limit
	if _, ok := allowedLimits[limit]; !ok {
		limit = defaultLimit
	}
	return page, limit
}
