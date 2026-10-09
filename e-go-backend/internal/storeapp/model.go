package storeapp

import (
	"net/http"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// Application statuses.
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// Store statuses.
const (
	StorePendingActivation = "pending_activation"
	StoreActive            = "active"
)

// Payment statuses (subset of the payment_status enum we act on).
const (
	PaymentPending    = "pending"
	PaymentProcessing = "processing"
	PaymentPaid       = "paid"
	PaymentFailed     = "failed"
	PaymentCancelled  = "cancelled"
)

// Business types.
const (
	BusinessIndividual = "individual"
	BusinessRegistered = "registered"
)

var (
	errInvalidBody = apperr.New(http.StatusBadRequest, "BAD_REQUEST", "Request body must be valid JSON")
	errNotFound    = apperr.New(http.StatusNotFound, "APPLICATION_NOT_FOUND", "Store application not found")
	errForbidden   = apperr.New(http.StatusForbidden, "FORBIDDEN", "You cannot access this store application")
	errSlugTaken   = apperr.New(http.StatusConflict, "STORE_SLUG_TAKEN", "That store URL is already taken")
	// errActiveApplication prevents duplicate in-flight applications.
	errActiveApplication = apperr.New(http.StatusConflict, "APPLICATION_IN_PROGRESS", "You already have a store application in progress")
	// errNotReviewable guards admin approve/reject on a non-pending application.
	errNotReviewable = apperr.New(http.StatusConflict, "APPLICATION_NOT_REVIEWABLE", "This application is no longer awaiting review")
	// errPaymentNotConfigured is returned when no real payment provider exists.
	errPaymentNotConfigured = apperr.New(http.StatusServiceUnavailable, "PAYMENT_NOT_CONFIGURED", "Online payment is not available yet — no payment provider is configured")
	// errPaymentNotRequired is returned when the fee is disabled.
	errPaymentNotRequired = apperr.New(http.StatusConflict, "PAYMENT_NOT_REQUIRED", "This application does not require an application fee")
	// errNotPayable guards payment initiation on a non-approved application.
	errNotPayable = apperr.New(http.StatusConflict, "APPLICATION_NOT_PAYABLE", "This application is not awaiting payment")
	// errNoActiveStore is returned when a caller has no active store to operate.
	errNoActiveStore = apperr.New(http.StatusForbidden, "NO_ACTIVE_STORE", "You do not have an active store")
	// errStoreNotActive is returned when a store exists but is not active.
	errStoreNotActive = apperr.New(http.StatusForbidden, "STORE_NOT_ACTIVE", "This store is not active")
	// errNotActivated is returned when activation is attempted before requirements are met.
	errNotActivated = apperr.New(http.StatusConflict, "STORE_NOT_READY", "The store is not ready to activate")
	// errNoLetter is returned when an approval letter is requested for an application not yet approved.
	errNoLetter = apperr.New(http.StatusConflict, "LETTER_NOT_AVAILABLE", "The approval letter is not available yet")
)

// Application is a store application (store_applications table) plus a few
// joined display fields.
type Application struct {
	ID                 string
	ApplicationNumber  string
	ApplicantUserID    string
	ApplicantName      string
	ApplicantEmail     string
	ApplicantStatus    string
	ApplicantCreatedAt time.Time
	PaymentStatus      string

	StoreName   string
	StoreSlug   string
	Description string
	CategoryID  *string
	Category    string

	LogoURL   string
	BannerURL string

	BusinessEmail string
	ContactPhone  string
	SupportPhone  string
	WebsiteURL    string

	Country     string
	Division    string
	District    string
	FullAddress string
	PostalCode  string

	BusinessType           string
	RegisteredBusinessName string
	TradeLicenseNumber     string

	Status           string
	RejectionReason  string
	AdminReviewNotes string
	ReviewedBy       *string
	ReviewedAt       *time.Time
	SubmittedAt      time.Time
	UpdatedAt        time.Time
}

// ApplicationParams is the validated, normalized create payload.
type ApplicationParams struct {
	ApplicationNumber      string
	ApplicantUserID        string
	StoreName              string
	StoreSlug              string
	Description            string
	CategoryID             *string
	LogoURL                string
	BannerURL              string
	BusinessEmail          string
	ContactPhone           string
	SupportPhone           string
	WebsiteURL             string
	Country                string
	Division               string
	District               string
	FullAddress            string
	PostalCode             string
	BusinessType           string
	RegisteredBusinessName string
	TradeLicenseNumber     string
}

// Store is the operational storefront (stores table).
type Store struct {
	ID            string
	VendorID      string
	ApplicationID *string
	Name          string
	Slug          string
	Description   string
	Category      string
	LogoURL       string
	BannerURL     string
	Status        string
	ActivatedAt   *time.Time
	CreatedAt     time.Time
	// Role is the caller's membership role (populated for my-stores only).
	Role string
}

// StoreContext is the resolved scope for a vendor request.
type StoreContext struct {
	StoreID  string
	VendorID string
	Name     string
	Status   string
	Role     string
}

// Payment is one application-fee payment (application_payments table).
type Payment struct {
	ID                    string
	ApplicationID         string
	UserID                string
	Amount                string
	Currency              string
	Provider              string
	Reference             string
	ProviderTransactionID *string
	Status                string
	VerifiedAt            *time.Time
	CreatedAt             time.Time
}

// PaymentParams is the validated payload for creating a payment record.
type PaymentParams struct {
	ApplicationID string
	UserID        string
	Amount        string
	Currency      string
	Provider      string
	Reference     string
}

// PaymentAdminRow is a payment row for the admin payments list.
type PaymentAdminRow struct {
	Payment
	ApplicationNumber string
	StoreName         string
	ApplicantName     string
	ApplicantEmail    string
}

// AdminListQuery is the normalized filter for admin application/store/payment lists.
type AdminListQuery struct {
	Page          int
	Limit         int
	Status        string
	PaymentStatus string
	Search        string
	Sort          string // submitted_at_desc | submitted_at_asc
}

// AuditEntry is one application history row.
type AuditEntry struct {
	Action    string
	ActorName string
	CreatedAt time.Time
}
