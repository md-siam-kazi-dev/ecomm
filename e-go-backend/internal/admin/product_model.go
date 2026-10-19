package admin

import "time"

// ReviewQuery filters the admin product review queue.
type ReviewQuery struct {
	// Status is "", "all", "pending", "rejected" or "changes_requested".
	Status string
	Search string
	Page   int
	Limit  int
}

// ReviewProduct is a product row in the admin review queue.
type ReviewProduct struct {
	ID             string
	Name           string
	VendorName     string
	Status         string
	ApprovalStatus string
	Price          float64
	Stock          int64
	ImageURL       string
	ReviewNote     string
	SubmittedAt    *time.Time
	ReviewedAt     *time.Time
	CreatedAt      time.Time
}

// ReviewImage is one product image in the review detail.
type ReviewImage struct {
	URL       string
	Alt       string
	IsPrimary bool
	SortOrder int
}

// ReviewProductDetail is the full product view for moderation.
type ReviewProductDetail struct {
	ID               string
	Name             string
	Slug             string
	Description      string
	ShortDescription string
	Status           string
	ApprovalStatus   string
	Category         string
	Brand            string
	SKU              string
	Price            float64
	SalePrice        *float64
	Stock            int64
	LowStock         int
	Images           []ReviewImage
	Attributes       map[string]string
	ReviewNote       string
	VendorName       string
	SubmittedAt      *time.Time
	ApprovedAt       *time.Time
	ReviewedAt       *time.Time
	CreatedAt        time.Time
}

// Decision is an admin moderation action.
type Decision string

const (
	DecisionApprove Decision = "approve"
	DecisionReject  Decision = "reject"
	DecisionChanges Decision = "changes"
)
