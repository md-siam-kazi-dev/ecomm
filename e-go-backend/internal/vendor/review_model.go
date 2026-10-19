package vendor

import (
	"net/http"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// errReviewNotFound is returned when a review does not exist or is not on one
// of the vendor's products.
var errReviewNotFound = apperr.New(http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found")

// errInvalidReply is returned for an empty or oversized vendor reply.
var errInvalidReply = apperr.New(
	http.StatusUnprocessableEntity, "VALIDATION_ERROR",
	"Enter a reply between 1 and 1000 characters",
)

// VendorReviewRecord is one review on the vendor's products.
type VendorReviewRecord struct {
	ID              string
	ProductID       string
	ProductName     string
	ProductSlug     string
	ProductImage    string
	Author          string
	OrderNumber     string
	Rating          int
	Title           string
	Comment         string
	Images          []string
	CreatedAt       time.Time
	VendorReply     string
	VendorRepliedAt *time.Time
}

// ReviewSummaryRecord is the vendor's aggregate star feedback.
type ReviewSummaryRecord struct {
	Total        int64
	Distribution [5]int64 // index 0 => 1 star ... index 4 => 5 stars
	Unreplied    int64
}
