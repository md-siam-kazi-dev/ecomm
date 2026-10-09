package storeapp

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/imgbb"
	"github.com/md-siam-kazi/e-go-backend/pkg/payment"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

const (
	defaultLimit = 20
	maxSearchLen = 100
	maxNoteLen   = 1000
	// maxApplicationImageBytes caps an uploaded logo/banner (5 MB).
	maxApplicationImageBytes = 5 << 20
)

var (
	// errInvalidImage is returned for a bad upload.
	errInvalidImage = apperr.New(http.StatusBadRequest, "INVALID_IMAGE", "Attach a PNG, JPG, WebP or GIF image (max 5 MB)")
	// errImageUnavailable is returned when image hosting is not configured.
	errImageUnavailable = apperr.New(http.StatusServiceUnavailable, "IMAGE_HOSTING_UNAVAILABLE", "Image uploads are not configured")
)

// ImageUploader hosts an image and returns its public URL.
type ImageUploader interface {
	Upload(ctx context.Context, filename string, data []byte) (string, error)
}

var allowedLimits = map[int]struct{}{10: {}, 20: {}, 50: {}, 100: {}}

var (
	allowedApplicationStatus = map[string]struct{}{StatusPending: {}, StatusApproved: {}, StatusRejected: {}}
	allowedStoreStatus       = map[string]struct{}{StorePendingActivation: {}, StoreActive: {}, "suspended": {}, "closed": {}}
	allowedPaymentStatus     = map[string]struct{}{
		PaymentPending: {}, PaymentProcessing: {}, PaymentPaid: {},
		PaymentFailed: {}, PaymentCancelled: {}, "partially_refunded": {}, "refunded": {},
	}
)

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Config carries the trusted server-side fee settings.
type Config struct {
	FeeEnabled   bool
	FeeAmount    string
	FeeCurrency  string
	ProviderName string
	PlatformName string
}

// Service owns the store-application business logic (Handler → Service → Repository).
type Service struct {
	repo     StoreRepository
	registry *payment.Registry
	provider payment.Provider
	cfg      Config
	uploader ImageUploader
}

// NewService builds the service, resolving the configured provider (falling back
// to the not-configured placeholder).
func NewService(repo StoreRepository, registry *payment.Registry, cfg Config) *Service {
	provider := payment.Provider(payment.NotConfigured{})
	if registry != nil {
		if p, err := registry.Get(cfg.ProviderName); err == nil {
			provider = p
		}
	}
	return &Service{repo: repo, registry: registry, provider: provider, cfg: cfg}
}

func (s *Service) feeInfo() FeeInfo {
	return FeeInfo{
		Enabled:            s.cfg.FeeEnabled,
		Amount:             s.cfg.FeeAmount,
		Currency:           s.cfg.FeeCurrency,
		ProviderConfigured: s.provider != nil && s.provider.Name() != "",
	}
}

func invalid(msg string) *apperr.Error {
	return apperr.New(http.StatusUnprocessableEntity, "VALIDATION_ERROR", msg)
}

// UploadImage hosts a store logo/banner and returns its public URL.
func (s *Service) UploadImage(ctx context.Context, filename string, data []byte) (string, error) {
	if s.uploader == nil || len(data) == 0 || len(data) > maxApplicationImageBytes {
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

// Submit validates and creates a pending store application.
func (s *Service) Submit(ctx context.Context, userID string, req SubmitRequest) (ApplicationDetail, error) {
	params, err := validateSubmit(userID, req)
	if err != nil {
		return ApplicationDetail{}, err
	}

	active, err := s.repo.HasActiveApplication(ctx, userID)
	if err != nil {
		return ApplicationDetail{}, err
	}
	if active {
		return ApplicationDetail{}, errActiveApplication
	}

	taken, err := s.repo.SlugTaken(ctx, params.StoreSlug)
	if err != nil {
		return ApplicationDetail{}, err
	}
	if taken {
		return ApplicationDetail{}, errSlugTaken
	}

	params.ApplicationNumber = "SA-" + time.Now().UTC().Format("20060102") + "-" + randomCode(6)
	app, err := s.repo.CreateApplication(ctx, params)
	if err != nil {
		return ApplicationDetail{}, err
	}
	return toApplicationDetail(app, nil, s.feeInfo(), nil), nil
}

// MyApplications lists the caller's applications.
func (s *Service) MyApplications(ctx context.Context, userID string) ([]ApplicationItem, error) {
	apps, err := s.repo.ApplicationsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]ApplicationItem, 0, len(apps))
	for _, a := range apps {
		items = append(items, toApplicationItem(a, a.PaymentStatus))
	}
	return items, nil
}

// Detail returns an application the caller owns (or any, for an admin).
func (s *Service) Detail(ctx context.Context, userID, id string, isAdmin bool) (ApplicationDetail, error) {
	app, err := s.repo.ApplicationByID(ctx, id)
	if err != nil {
		return ApplicationDetail{}, err
	}
	if !isAdmin && app.ApplicantUserID != userID {
		return ApplicationDetail{}, errForbidden
	}
	store, err := s.repo.StoreForApplication(ctx, id)
	if err != nil {
		return ApplicationDetail{}, err
	}
	p, err := s.repo.LatestPayment(ctx, id)
	if err != nil {
		return ApplicationDetail{}, err
	}
	return toApplicationDetail(app, store, s.feeInfo(), p), nil
}

// AdminDetail returns the full admin view including internal notes and history.
func (s *Service) AdminDetail(ctx context.Context, id string) (AdminApplicationDetail, error) {
	app, err := s.repo.ApplicationByID(ctx, id)
	if err != nil {
		return AdminApplicationDetail{}, err
	}
	store, err := s.repo.StoreForApplication(ctx, id)
	if err != nil {
		return AdminApplicationDetail{}, err
	}
	p, err := s.repo.LatestPayment(ctx, id)
	if err != nil {
		return AdminApplicationDetail{}, err
	}

	history, err := s.repo.AuditForApplication(ctx, id)
	if err != nil {
		return AdminApplicationDetail{}, err
	}
	items := make([]HistoryItem, 0, len(history))
	for _, h := range history {
		items = append(items, HistoryItem{
			Action:    h.Action,
			ActorName: h.ActorName,
			CreatedAt: h.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return AdminApplicationDetail{
		ApplicationDetail: toApplicationDetail(app, store, s.feeInfo(), p),
		ApplicantUserID:   app.ApplicantUserID,
		ApplicantAccount: ApplicantAccount{
			ID:        app.ApplicantUserID,
			Email:     app.ApplicantEmail,
			Name:      app.ApplicantName,
			Status:    app.ApplicantStatus,
			CreatedAt: app.ApplicantCreatedAt.UTC().Format(time.RFC3339),
		},
		AdminReviewNotes: app.AdminReviewNotes,
		ReviewedBy:       app.ReviewedBy,
		History:          items,
	}, nil
}

// AdminListApplications returns a page of applications for the review queue.
func (s *Service) AdminListApplications(ctx context.Context, q AdminListQuery) ([]ApplicationItem, response.Pagination, error) {
	q = normalizeQuery(q, allowedApplicationStatus)
	apps, total, err := s.repo.ListApplications(ctx, q)
	if err != nil {
		return nil, response.Pagination{}, err
	}
	items := make([]ApplicationItem, 0, len(apps))
	for _, a := range apps {
		items = append(items, toApplicationItem(a, a.PaymentStatus))
	}
	return items, response.Pagination{Page: q.Page, Limit: q.Limit, Total: total}, nil
}

// Approve approves a pending application (idempotency/state enforced in the repo).
func (s *Service) Approve(ctx context.Context, adminID, id string) (AdminApplicationDetail, error) {
	if _, err := s.repo.Approve(ctx, id, adminID, s.cfg.FeeEnabled); err != nil {
		return AdminApplicationDetail{}, err
	}
	return s.AdminDetail(ctx, id)
}

// Reject rejects a pending application with a required reason.
func (s *Service) Reject(ctx context.Context, adminID, id, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return invalid("A rejection reason is required")
	}
	if len(reason) > maxNoteLen {
		reason = reason[:maxNoteLen]
	}
	return s.repo.Reject(ctx, id, adminID, reason)
}

// UpdateNotes stores an internal admin note.
func (s *Service) UpdateNotes(ctx context.Context, id, note string) error {
	note = strings.TrimSpace(note)
	if len(note) > maxNoteLen {
		note = note[:maxNoteLen]
	}
	return s.repo.UpdateNotes(ctx, id, note)
}

// MyStores lists the active stores the caller can access.
func (s *Service) MyStores(ctx context.Context, userID string) ([]StoreItem, error) {
	stores, err := s.repo.StoresForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]StoreItem, 0, len(stores))
	for _, st := range stores {
		items = append(items, toStoreItem(st, st.Role))
	}
	return items, nil
}

// AdminListStores returns a page of stores.
func (s *Service) AdminListStores(ctx context.Context, q AdminListQuery) ([]StoreItem, response.Pagination, error) {
	q = normalizeQuery(q, allowedStoreStatus)
	stores, total, err := s.repo.ListStores(ctx, q)
	if err != nil {
		return nil, response.Pagination{}, err
	}
	items := make([]StoreItem, 0, len(stores))
	for _, st := range stores {
		items = append(items, toStoreItem(st, "owner"))
	}
	return items, response.Pagination{Page: q.Page, Limit: q.Limit, Total: total}, nil
}

// AdminListPayments returns a page of application payments.
func (s *Service) AdminListPayments(ctx context.Context, q AdminListQuery) ([]PaymentAdminItem, response.Pagination, error) {
	q = normalizeQuery(q, allowedPaymentStatus)
	rows, total, err := s.repo.ListPayments(ctx, q)
	if err != nil {
		return nil, response.Pagination{}, err
	}
	items := make([]PaymentAdminItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, toPaymentAdminItem(r))
	}
	return items, response.Pagination{Page: q.Page, Limit: q.Limit, Total: total}, nil
}

// InitiatePayment creates (or reuses) an application-fee payment and asks the
// provider to start a session. When no provider is configured it returns an
// explicit "not configured" error and creates no record.
func (s *Service) InitiatePayment(ctx context.Context, userID, applicationID string) (PaymentInitResponse, error) {
	app, err := s.repo.ApplicationByID(ctx, applicationID)
	if err != nil {
		return PaymentInitResponse{}, err
	}
	if app.ApplicantUserID != userID {
		return PaymentInitResponse{}, errForbidden
	}
	if app.Status != StatusApproved {
		return PaymentInitResponse{}, errNotPayable
	}
	if !s.cfg.FeeEnabled {
		return PaymentInitResponse{}, errPaymentNotRequired
	}
	if s.provider == nil || s.provider.Name() == "" {
		return PaymentInitResponse{}, errPaymentNotConfigured
	}

	// Reuse an in-flight payment so repeated clicks don't create duplicates.
	latest, err := s.repo.LatestPayment(ctx, applicationID)
	if err != nil {
		return PaymentInitResponse{}, err
	}
	if latest != nil && (latest.Status == PaymentPending || latest.Status == PaymentProcessing) {
		init, _ := s.provider.Initiate(ctx, payment.InitRequest{
			Reference: latest.Reference, Amount: latest.Amount, Currency: latest.Currency,
		})
		return PaymentInitResponse{
			Payment:            toPaymentItem(*latest),
			ProviderConfigured: true,
			RedirectURL:        init.RedirectURL,
			ClientPayload:      init.ClientPayload,
		}, nil
	}

	ref := "PAY-" + randomCode(14)
	record, err := s.repo.CreatePayment(ctx, PaymentParams{
		ApplicationID: applicationID,
		UserID:        userID,
		Amount:        s.cfg.FeeAmount,
		Currency:      s.cfg.FeeCurrency,
		Provider:      s.provider.Name(),
		Reference:     ref,
	})
	if err != nil {
		return PaymentInitResponse{}, err
	}

	init, err := s.provider.Initiate(ctx, payment.InitRequest{
		Reference: ref,
		Amount:    s.cfg.FeeAmount,
		Currency:  s.cfg.FeeCurrency,
		Metadata:  map[string]string{"application_id": applicationID},
	})
	if err != nil {
		if errors.Is(err, payment.ErrNotConfigured) {
			return PaymentInitResponse{}, errPaymentNotConfigured
		}
		return PaymentInitResponse{}, apperr.ErrInternal.WithCause(err)
	}

	return PaymentInitResponse{
		Payment:            toPaymentItem(record),
		ProviderConfigured: true,
		RedirectURL:        init.RedirectURL,
		ClientPayload:      init.ClientPayload,
	}, nil
}

// Payments lists the payments for an application the caller owns.
func (s *Service) Payments(ctx context.Context, userID, applicationID string) ([]PaymentItem, error) {
	app, err := s.repo.ApplicationByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if app.ApplicantUserID != userID {
		return nil, errForbidden
	}
	records, err := s.repo.PaymentsForApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	items := make([]PaymentItem, 0, len(records))
	for _, p := range records {
		items = append(items, toPaymentItem(p))
	}
	return items, nil
}

// HandleWebhook verifies a provider callback server-side and activates the
// store on a verified payment. Idempotent for repeated callbacks.
func (s *Service) HandleWebhook(ctx context.Context, providerName string, req payment.VerifyRequest) (PaymentItem, error) {
	if s.registry == nil {
		return PaymentItem{}, errPaymentNotConfigured
	}
	provider, err := s.registry.Get(providerName)
	if err != nil || provider.Name() == "" {
		return PaymentItem{}, errPaymentNotConfigured
	}

	result, err := provider.Verify(ctx, req)
	if err != nil {
		if errors.Is(err, payment.ErrNotConfigured) {
			return PaymentItem{}, errPaymentNotConfigured
		}
		return PaymentItem{}, apperr.ErrInternal.WithCause(err)
	}
	if !result.Paid {
		return PaymentItem{}, invalid("Payment could not be verified")
	}

	record, err := s.repo.PaymentByReference(ctx, result.Reference)
	if err != nil {
		return PaymentItem{}, err
	}
	updated, _, err := s.repo.MarkPaid(ctx, result.Reference, result.ProviderTransactionID)
	if err != nil {
		return PaymentItem{}, err
	}
	// Activation is idempotent; safe to attempt on every verified callback.
	if _, err := s.repo.ActivateForApplication(ctx, record.ApplicationID); err != nil {
		return PaymentItem{}, err
	}
	return toPaymentItem(updated), nil
}

// --- helpers ---

func normalizeQuery(q AdminListQuery, allowed map[string]struct{}) AdminListQuery {
	if q.Page < 1 {
		q.Page = 1
	}
	if _, ok := allowedLimits[q.Limit]; !ok {
		q.Limit = defaultLimit
	}
	q.Status = strings.ToLower(strings.TrimSpace(q.Status))
	if _, ok := allowed[q.Status]; !ok {
		q.Status = ""
	}
	q.PaymentStatus = strings.ToLower(strings.TrimSpace(q.PaymentStatus))
	if _, ok := allowedPaymentStatus[q.PaymentStatus]; !ok {
		q.PaymentStatus = ""
	}
	q.Search = strings.TrimSpace(q.Search)
	if len(q.Search) > maxSearchLen {
		q.Search = q.Search[:maxSearchLen]
	}
	if q.Sort != "submitted_at_asc" {
		q.Sort = "submitted_at_desc"
	}
	return q
}

func validateSubmit(userID string, req SubmitRequest) (ApplicationParams, error) {
	p := ApplicationParams{ApplicantUserID: userID}

	p.StoreName = strings.TrimSpace(req.StoreName)
	if l := len(p.StoreName); l < 2 || l > 255 {
		return p, invalid("Store name must be between 2 and 255 characters")
	}

	slug := normalizeSlug(req.StoreSlug)
	if slug == "" {
		slug = normalizeSlug(req.StoreName)
	}
	if slug == "" {
		return p, invalid("Store URL is required")
	}
	if len(slug) < 3 || len(slug) > 255 {
		return p, invalid("Store URL must be at least 3 characters")
	}
	if !slugRe.MatchString(slug) {
		return p, invalid("Store URL may only contain lowercase letters, numbers and hyphens")
	}
	p.StoreSlug = slug

	p.Description = strings.TrimSpace(req.Description)
	if p.Description == "" {
		return p, invalid("Store description is required")
	}
	if len(p.Description) > 5000 {
		p.Description = p.Description[:5000]
	}

	categoryID := strings.TrimSpace(req.CategoryID)
	if categoryID == "" {
		return p, invalid("Please choose a store category")
	}
	p.CategoryID = &categoryID

	p.LogoURL = strings.TrimSpace(req.LogoURL)
	p.BannerURL = strings.TrimSpace(req.BannerURL)

	email := strings.ToLower(strings.TrimSpace(req.BusinessEmail))
	if email == "" {
		return p, invalid("Business email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil || len(email) > 255 {
		return p, invalid("Enter a valid business email")
	}
	p.BusinessEmail = email

	p.ContactPhone = strings.TrimSpace(req.ContactPhone)
	if p.ContactPhone == "" {
		return p, invalid("Contact phone is required")
	}
	if len(p.ContactPhone) > 30 {
		return p, invalid("Contact phone is too long")
	}
	p.SupportPhone = strings.TrimSpace(req.SupportPhone)

	if site := strings.TrimSpace(req.WebsiteURL); site != "" {
		u, err := url.Parse(site)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return p, invalid("Enter a valid website URL (http:// or https://)")
		}
		p.WebsiteURL = site
	}

	p.Country = strings.TrimSpace(req.Country)
	if p.Country == "" {
		p.Country = "Bangladesh"
	}
	p.Division = strings.TrimSpace(req.Division)
	p.District = strings.TrimSpace(req.District)
	p.FullAddress = strings.TrimSpace(req.FullAddress)
	p.PostalCode = strings.TrimSpace(req.PostalCode)
	if p.Division == "" {
		return p, invalid("Division/state is required")
	}
	if p.District == "" {
		return p, invalid("District/city is required")
	}
	if p.FullAddress == "" {
		return p, invalid("Full address is required")
	}

	bt := strings.ToLower(strings.TrimSpace(req.BusinessType))
	if bt != BusinessIndividual && bt != BusinessRegistered {
		return p, invalid("Choose a business type")
	}
	p.BusinessType = bt
	p.RegisteredBusinessName = strings.TrimSpace(req.RegisteredBusinessName)
	if bt == BusinessRegistered && p.RegisteredBusinessName == "" {
		return p, invalid("Registered business name is required")
	}
	p.TradeLicenseNumber = strings.TrimSpace(req.TradeLicenseNumber)

	if !req.AgreedToTerms {
		return p, invalid("Confirm the information is accurate and agree to the vendor terms")
	}
	return p, nil
}

func normalizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_' || r == '.' || r == '/':
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

const codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

func randomCode(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		for i := range b {
			b[i] = byte(time.Now().UnixNano() >> (i % 8))
		}
	}
	for i := range b {
		b[i] = codeAlphabet[int(b[i])%len(codeAlphabet)]
	}
	return string(b)
}
