package storeapp

import "time"

// SubmitRequest is the customer's store application payload.
type SubmitRequest struct {
	StoreName   string `json:"store_name"`
	StoreSlug   string `json:"store_slug"`
	Description string `json:"description"`
	// Category is the store category: a category slug or id, resolved to
	// categories.id server-side.
	CategoryID             string `json:"category"`
	LogoURL                string `json:"logo_url"`
	BannerURL              string `json:"banner_url"`
	BusinessEmail          string `json:"business_email"`
	ContactPhone           string `json:"contact_phone"`
	SupportPhone           string `json:"support_phone"`
	WebsiteURL             string `json:"website_url"`
	Country                string `json:"country"`
	Division               string `json:"division"`
	District               string `json:"district"`
	FullAddress            string `json:"full_address"`
	PostalCode             string `json:"postal_code"`
	BusinessType           string `json:"business_type"`
	RegisteredBusinessName string `json:"registered_business_name"`
	TradeLicenseNumber     string `json:"trade_license_number"`
	AgreedToTerms          bool   `json:"agreed_to_terms"`
}

// RejectRequest carries the admin rejection reason.
type RejectRequest struct {
	Reason string `json:"reason"`
}

// NoteRequest carries an internal admin note update.
type NoteRequest struct {
	Note string `json:"note"`
}

// ApplicationItem is a list row.
type ApplicationItem struct {
	ID                string `json:"id"`
	ApplicationNumber string `json:"application_number"`
	ApplicantName     string `json:"applicant_name"`
	ApplicantEmail    string `json:"applicant_email"`
	StoreName         string `json:"store_name"`
	Category          string `json:"category"`
	Status            string `json:"status"`
	PaymentStatus     string `json:"payment_status"`
	SubmittedAt       string `json:"submitted_at"`
}

// StoreItem is a store the caller can access.
type StoreItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	LogoURL     string  `json:"logo_url"`
	Status      string  `json:"status"`
	Role        string  `json:"role"`
	ActivatedAt *string `json:"activated_at,omitempty"`
}

// PaymentItem describes one application-fee payment.
type PaymentItem struct {
	Reference  string  `json:"reference"`
	Amount     string  `json:"amount"`
	Currency   string  `json:"currency"`
	Status     string  `json:"status"`
	Provider   string  `json:"provider,omitempty"`
	CreatedAt  string  `json:"created_at"`
	VerifiedAt *string `json:"verified_at,omitempty"`
}

// FeeInfo is the application-fee configuration as seen by the applicant.
type FeeInfo struct {
	Enabled            bool   `json:"enabled"`
	Amount             string `json:"amount"`
	Currency           string `json:"currency"`
	ProviderConfigured bool   `json:"provider_configured"`
}

// ApplicationDetail is the applicant-facing detail view (no internal notes).
type ApplicationDetail struct {
	ID                string `json:"id"`
	ApplicationNumber string `json:"application_number"`
	StoreName         string `json:"store_name"`
	StoreSlug         string `json:"store_slug"`
	Description       string `json:"description"`
	Category          string `json:"category"`
	LogoURL           string `json:"logo_url"`
	BannerURL         string `json:"banner_url"`

	BusinessEmail string `json:"business_email"`
	ContactPhone  string `json:"contact_phone"`
	SupportPhone  string `json:"support_phone"`
	WebsiteURL    string `json:"website_url"`

	Country     string `json:"country"`
	Division    string `json:"division"`
	District    string `json:"district"`
	FullAddress string `json:"full_address"`
	PostalCode  string `json:"postal_code"`

	BusinessType           string `json:"business_type"`
	RegisteredBusinessName string `json:"registered_business_name"`
	TradeLicenseNumber     string `json:"trade_license_number"`

	Status                  string  `json:"status"`
	RejectionReason         string  `json:"rejection_reason,omitempty"`
	SubmittedAt             string  `json:"submitted_at"`
	ReviewedAt              *string `json:"reviewed_at,omitempty"`
	ApprovedAt              *string `json:"approved_at,omitempty"`
	ApprovalLetterAvailable bool    `json:"approval_letter_available"`

	Store   *StoreItem   `json:"store,omitempty"`
	Fee     *FeeInfo     `json:"fee,omitempty"`
	Payment *PaymentItem `json:"payment,omitempty"`
	// NextAction drives the applicant UI: apply | await_review | pay_fee |
	// payment_processing | open_dashboard | rejected | suspended | none.
	NextAction string `json:"next_action"`
}

// ApplicantAccount is the applicant's account summary (admin view).
type ApplicantAccount struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// HistoryItem is one audit/history entry.
type HistoryItem struct {
	Action    string `json:"action"`
	ActorName string `json:"actor_name"`
	CreatedAt string `json:"created_at"`
}

// AdminApplicationDetail is the full admin view including internal fields.
type AdminApplicationDetail struct {
	ApplicationDetail
	ApplicantUserID  string           `json:"applicant_user_id"`
	ApplicantAccount ApplicantAccount `json:"applicant_account"`
	AdminReviewNotes string           `json:"admin_review_notes,omitempty"`
	ReviewedBy       *string          `json:"reviewed_by,omitempty"`
	History          []HistoryItem    `json:"history"`
}

// PaymentInitResponse is returned when an application-fee payment is started.
type PaymentInitResponse struct {
	Payment            PaymentItem       `json:"payment"`
	ProviderConfigured bool              `json:"provider_configured"`
	RedirectURL        string            `json:"redirect_url,omitempty"`
	ClientPayload      map[string]string `json:"client_payload,omitempty"`
}

// PaymentAdminItem is a payment row for the admin payments list.
type PaymentAdminItem struct {
	ID                    string  `json:"id"`
	Reference             string  `json:"reference"`
	ApplicationNumber     string  `json:"application_number"`
	StoreName             string  `json:"store_name"`
	ApplicantName         string  `json:"applicant_name"`
	ApplicantEmail        string  `json:"applicant_email"`
	Amount                string  `json:"amount"`
	Currency              string  `json:"currency"`
	Status                string  `json:"status"`
	Provider              string  `json:"provider,omitempty"`
	ProviderTransactionID *string `json:"provider_transaction_id,omitempty"`
	CreatedAt             string  `json:"created_at"`
	VerifiedAt            *string `json:"verified_at,omitempty"`
}

// ApprovalLetterData is the trusted data used to render the approval letter.
type ApprovalLetterData struct {
	PlatformName      string
	ApplicationNumber string
	ApplicantName     string
	StoreName         string
	StoreCategory     string
	ApprovalDate      string
	FeeEnabled        bool
	FeeAmount         string
	FeeCurrency       string
}

func toPaymentAdminItem(r PaymentAdminRow) PaymentAdminItem {
	item := PaymentAdminItem{
		ID:                    r.ID,
		Reference:             r.Reference,
		ApplicationNumber:     r.ApplicationNumber,
		StoreName:             r.StoreName,
		ApplicantName:         r.ApplicantName,
		ApplicantEmail:        r.ApplicantEmail,
		Amount:                r.Amount,
		Currency:              r.Currency,
		Status:                r.Status,
		Provider:              r.Provider,
		ProviderTransactionID: r.ProviderTransactionID,
		CreatedAt:             r.CreatedAt.UTC().Format(time.RFC3339),
	}
	if r.VerifiedAt != nil {
		v := r.VerifiedAt.UTC().Format(time.RFC3339)
		item.VerifiedAt = &v
	}
	return item
}

func toApplicationItem(a Application, paymentStatus string) ApplicationItem {
	return ApplicationItem{
		ID:                a.ID,
		ApplicationNumber: a.ApplicationNumber,
		ApplicantName:     a.ApplicantName,
		ApplicantEmail:    a.ApplicantEmail,
		StoreName:         a.StoreName,
		Category:          a.Category,
		Status:            a.Status,
		PaymentStatus:     paymentStatus,
		SubmittedAt:       a.SubmittedAt.UTC().Format(time.RFC3339),
	}
}

func toStoreItem(s Store, role string) StoreItem {
	item := StoreItem{
		ID:      s.ID,
		Name:    s.Name,
		Slug:    s.Slug,
		LogoURL: s.LogoURL,
		Status:  s.Status,
		Role:    role,
	}
	if s.ActivatedAt != nil {
		v := s.ActivatedAt.UTC().Format(time.RFC3339)
		item.ActivatedAt = &v
	}
	return item
}

func toPaymentItem(p Payment) PaymentItem {
	item := PaymentItem{
		Reference: p.Reference,
		Amount:    p.Amount,
		Currency:  p.Currency,
		Status:    p.Status,
		Provider:  p.Provider,
		CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339),
	}
	if p.VerifiedAt != nil {
		v := p.VerifiedAt.UTC().Format(time.RFC3339)
		item.VerifiedAt = &v
	}
	return item
}

// nextAction derives the applicant's next step from application/store/payment state.
func nextAction(a Application, store *Store, fee FeeInfo, payment *Payment) string {
	if a.Status == StatusRejected {
		return "rejected"
	}
	if store != nil && store.Status == "suspended" {
		return "suspended"
	}
	if a.Status == StatusPending {
		return "await_review"
	}
	// approved
	if store != nil && store.Status == StoreActive {
		return "open_dashboard"
	}
	if fee.Enabled {
		if payment != nil {
			switch payment.Status {
			case PaymentPaid:
				return "open_dashboard"
			case PaymentPending, PaymentProcessing:
				return "payment_processing"
			}
		}
		return "pay_fee"
	}
	return "open_dashboard"
}

func toApplicationDetail(a Application, store *Store, fee FeeInfo, payment *Payment) ApplicationDetail {
	d := ApplicationDetail{
		ID:                      a.ID,
		ApplicationNumber:       a.ApplicationNumber,
		StoreName:               a.StoreName,
		StoreSlug:               a.StoreSlug,
		Description:             a.Description,
		Category:                a.Category,
		LogoURL:                 a.LogoURL,
		BannerURL:               a.BannerURL,
		BusinessEmail:           a.BusinessEmail,
		ContactPhone:            a.ContactPhone,
		SupportPhone:            a.SupportPhone,
		WebsiteURL:              a.WebsiteURL,
		Country:                 a.Country,
		Division:                a.Division,
		District:                a.District,
		FullAddress:             a.FullAddress,
		PostalCode:              a.PostalCode,
		BusinessType:            a.BusinessType,
		RegisteredBusinessName:  a.RegisteredBusinessName,
		TradeLicenseNumber:      a.TradeLicenseNumber,
		Status:                  a.Status,
		RejectionReason:         a.RejectionReason,
		SubmittedAt:             a.SubmittedAt.UTC().Format(time.RFC3339),
		ApprovalLetterAvailable: a.Status == StatusApproved,
		NextAction:              nextAction(a, store, fee, payment),
	}
	if a.ReviewedAt != nil {
		v := a.ReviewedAt.UTC().Format(time.RFC3339)
		d.ReviewedAt = &v
	}
	if a.Status == StatusApproved && a.ReviewedAt != nil {
		v := a.ReviewedAt.UTC().Format(time.RFC3339)
		d.ApprovedAt = &v
	}
	if store != nil {
		si := toStoreItem(*store, "owner")
		d.Store = &si
	}
	d.Fee = &fee
	if payment != nil {
		pi := toPaymentItem(*payment)
		d.Payment = &pi
	}
	return d
}
