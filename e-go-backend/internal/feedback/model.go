package feedback

import (
	"net/http"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// Status values used for the logged order status check.
const orderStatusDelivered = "delivered"

// OrderItemEligibility is the server-resolved purchase context for a review.
// ProductID is empty when the underlying product row no longer exists.
type OrderItemEligibility struct {
	ProductID   string
	UserID      string
	OrderStatus string
}

// ReviewParams is a validated review ready to persist.
type ReviewParams struct {
	ProductID   string
	UserID      string
	OrderItemID string
	Rating      int
	Title       string
	Comment     string
	Images      []string
}

// ReviewRecord is a persisted review.
type ReviewRecord struct {
	ID          string
	ProductID   string
	UserID      string
	OrderItemID string
	Rating      int
	Title       string
	Comment     string
	Status      string
	Images      []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// QuestionRecord is a persisted product question.
type QuestionRecord struct {
	ID        string
	ProductID string
	UserID    string
	Question  string
	CreatedAt time.Time
}

var (
	errInvalidBody = apperr.New(http.StatusBadRequest, "BAD_REQUEST", "Request body must be valid JSON")

	errInvalidReview = apperr.New(
		http.StatusUnprocessableEntity, "VALIDATION_ERROR",
		"Enter a rating from 1 to 5 and a comment of at least 3 characters",
	)

	errOrderNotFound = apperr.New(http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")

	errOrderNotDelivered = apperr.New(
		http.StatusUnprocessableEntity, "ORDER_NOT_DELIVERED",
		"You can review this product once your order is delivered",
	)

	errProductUnavailable = apperr.New(
		http.StatusUnprocessableEntity, "PRODUCT_UNAVAILABLE",
		"This product is no longer available to review",
	)

	errReviewExists = apperr.New(
		http.StatusConflict, "REVIEW_EXISTS",
		"You have already reviewed this purchase",
	)

	errReviewNotFound = apperr.New(http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found")

	errProductNotFound = apperr.New(http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")

	errInvalidQuestion = apperr.New(
		http.StatusUnprocessableEntity, "VALIDATION_ERROR",
		"Enter a question between 5 and 1000 characters",
	)

	errInvalidImages = apperr.New(
		http.StatusUnprocessableEntity, "VALIDATION_ERROR",
		"Attach at most 4 photos to a review",
	)

	errInvalidImage = apperr.New(
		http.StatusBadRequest, "INVALID_IMAGE",
		"Attach a PNG, JPG, WebP or GIF image (max 5 MB)",
	)

	errImageUnavailable = apperr.New(
		http.StatusServiceUnavailable, "IMAGE_HOSTING_UNAVAILABLE",
		"Image uploads are not configured",
	)
)
