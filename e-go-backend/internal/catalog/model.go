package catalog

import (
	"net/http"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// errProductNotFound is returned when a slug does not match a live product.
var errProductNotFound = apperr.New(http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")

// ProductRecord is one catalog product row (with aggregates).
type ProductRecord struct {
	ID               string
	Slug             string
	Name             string
	ShortDescription string
	Description      string
	Price            float64
	CompareAtPrice   *float64
	SKU              string
	Stock            int64
	ImageURL         string
	ImageAlt         string
	Category         string
	CategorySlug     string
	Brand            string
	Rating           float64
	ReviewCount      int64
	SalesUnits       int64
	IsNew            bool
	Attributes       map[string]string
	SeoTitle         string
	SeoDescription   string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// CategoryRecord is a storefront category with its live product count.
type CategoryRecord struct {
	ID           string
	Slug         string
	Name         string
	Description  string
	ImageURL     string
	ProductCount int64
}

// ReviewRecord is a published review.
type ReviewRecord struct {
	ID               string
	ProductID        string
	ProductName      string
	ProductSlug      string
	Author           string
	Rating           int
	Title            string
	Comment          string
	VerifiedPurchase bool
	HelpfulCount     int64
	VendorReply      string
	VendorRepliedAt  *time.Time
	CreatedAt        time.Time
}

// RatingSummaryRecord is a product's published-review rating aggregate.
type RatingSummaryRecord struct {
	Total        int64
	Distribution [5]int64 // index 0 => 1 star ... index 4 => 5 stars
}

// AnswerRecord is a public answer to a product question.
type AnswerRecord struct {
	ID                     string
	Answer                 string
	IsOfficialVendorAnswer bool
	Author                 string
	CreatedAt              time.Time
}

// QuestionRecord is a published product question with its public answers.
type QuestionRecord struct {
	ID        string
	Question  string
	Author    string
	CreatedAt time.Time
	Answers   []AnswerRecord
}

// ListQuery is the normalized catalog list filter.
type ListQuery struct {
	Category  string
	Search    string
	MinPrice  *float64
	MaxPrice  *float64
	MinRating *float64
	InStock   bool
	OnSale    bool
	Sort      string
	Page      int
	Limit     int
}
