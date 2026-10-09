package storeapp

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

const pgUniqueViolation = "23505"

// StoreRepository is the persistence boundary for the store-application module.
type StoreRepository interface {
	CreateApplication(ctx context.Context, p ApplicationParams) (Application, error)
	SlugTaken(ctx context.Context, slug string) (bool, error)
	HasActiveApplication(ctx context.Context, userID string) (bool, error)
	ApplicationsForUser(ctx context.Context, userID string) ([]Application, error)
	ApplicationByID(ctx context.Context, id string) (Application, error)
	ListApplications(ctx context.Context, q AdminListQuery) ([]Application, int64, error)
	ApplicationPaymentStatus(ctx context.Context, applicationID string) (string, error)
	AuditForApplication(ctx context.Context, id string) ([]AuditEntry, error)

	Approve(ctx context.Context, id, adminID string, feeEnabled bool) (Application, error)
	Reject(ctx context.Context, id, adminID, reason string) error
	UpdateNotes(ctx context.Context, id, note string) error

	StoresForUser(ctx context.Context, userID string) ([]Store, error)
	ResolveStore(ctx context.Context, userID, storeID string) (StoreContext, error)
	StoreForApplication(ctx context.Context, applicationID string) (*Store, error)
	ListStores(ctx context.Context, q AdminListQuery) ([]Store, int64, error)

	CreatePayment(ctx context.Context, p PaymentParams) (Payment, error)
	PaymentsForApplication(ctx context.Context, applicationID string) ([]Payment, error)
	LatestPayment(ctx context.Context, applicationID string) (*Payment, error)
	PaymentByReference(ctx context.Context, ref string) (Payment, error)
	MarkPaid(ctx context.Context, ref, providerTxnID string) (Payment, bool, error)
	ActivateForApplication(ctx context.Context, applicationID string) (Store, error)
	ListPayments(ctx context.Context, q AdminListQuery) ([]PaymentAdminRow, int64, error)
}

// Repository persists store-application data in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

type scanner interface{ Scan(dest ...any) error }

const applicationColumns = `
	a.id, a.application_number, a.applicant_user_id,
	COALESCE(NULLIF(TRIM(CONCAT_WS(' ', p.first_name, p.last_name)), ''), split_part(u.email, '@', 1)) AS applicant_name,
	u.email,
	u.status::text, u.created_at,
	COALESCE((SELECT ap.status::text FROM application_payments ap WHERE ap.application_id = a.id ORDER BY ap.created_at DESC LIMIT 1), '') AS payment_status,
	a.store_name, a.store_slug, a.description, a.category_id, COALESCE(c.name, ''),
	COALESCE(a.logo_url, ''), COALESCE(a.banner_url, ''),
	a.business_email, a.contact_phone, COALESCE(a.support_phone, ''), COALESCE(a.website_url, ''),
	a.country, a.division, a.district, a.full_address, COALESCE(a.postal_code, ''),
	a.business_type::text, COALESCE(a.registered_business_name, ''), COALESCE(a.trade_license_number, ''),
	a.status::text, COALESCE(a.rejection_reason, ''), COALESCE(a.admin_review_notes, ''),
	a.reviewed_by, a.reviewed_at, a.submitted_at, a.updated_at`

const applicationFrom = `
	FROM store_applications a
	JOIN users u ON u.id = a.applicant_user_id
	LEFT JOIN user_profiles p ON p.user_id = u.id
	LEFT JOIN categories c ON c.id = a.category_id`

func scanApplication(s scanner) (Application, error) {
	var a Application
	err := s.Scan(
		&a.ID, &a.ApplicationNumber, &a.ApplicantUserID, &a.ApplicantName, &a.ApplicantEmail,
		&a.ApplicantStatus, &a.ApplicantCreatedAt, &a.PaymentStatus,
		&a.StoreName, &a.StoreSlug, &a.Description, &a.CategoryID, &a.Category,
		&a.LogoURL, &a.BannerURL,
		&a.BusinessEmail, &a.ContactPhone, &a.SupportPhone, &a.WebsiteURL,
		&a.Country, &a.Division, &a.District, &a.FullAddress, &a.PostalCode,
		&a.BusinessType, &a.RegisteredBusinessName, &a.TradeLicenseNumber,
		&a.Status, &a.RejectionReason, &a.AdminReviewNotes,
		&a.ReviewedBy, &a.ReviewedAt, &a.SubmittedAt, &a.UpdatedAt,
	)
	return a, err
}

func nullable(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

// CreateApplication inserts a new pending application.
func (r *Repository) CreateApplication(ctx context.Context, p ApplicationParams) (Application, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO store_applications (
			application_number, applicant_user_id, store_name, store_slug, description, category_id,
			logo_url, banner_url, business_email, contact_phone, support_phone, website_url,
			country, division, district, full_address, postal_code,
			business_type, registered_business_name, trade_license_number
		) VALUES (
			$1, $2, $3, $4, $5, (SELECT id FROM categories WHERE slug = $6 OR id::text = $6 LIMIT 1),
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17,
			$18::business_type, $19, $20
		) RETURNING id
	`,
		p.ApplicationNumber, p.ApplicantUserID, p.StoreName, p.StoreSlug, p.Description, nullablePtr(p.CategoryID),
		nullable(p.LogoURL), nullable(p.BannerURL), p.BusinessEmail, p.ContactPhone, nullable(p.SupportPhone), nullable(p.WebsiteURL),
		p.Country, p.Division, p.District, p.FullAddress, nullable(p.PostalCode),
		p.BusinessType, nullable(p.RegisteredBusinessName), nullable(p.TradeLicenseNumber),
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			if strings.Contains(pgErr.ConstraintName, "slug") {
				return Application{}, errSlugTaken.WithCause(err)
			}
			return Application{}, apperr.ErrConflict.WithCause(err)
		}
		return Application{}, apperr.ErrInternal.WithCause(err)
	}
	// Best-effort history entry; never blocks submission.
	_, _ = r.db.Exec(ctx, `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id)
		VALUES ($1, 'store_application.submitted', 'store_application', $2)`, p.ApplicantUserID, id)
	return r.ApplicationByID(ctx, id)
}

func nullablePtr(s *string) *string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	return s
}

// SlugTaken reports whether a store slug is already in use by an application or
// an existing store/vendor.
func (r *Repository) SlugTaken(ctx context.Context, slug string) (bool, error) {
	var taken bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM store_applications WHERE store_slug = $1 AND status IN ('pending', 'approved'))
		    OR EXISTS (SELECT 1 FROM stores WHERE slug = $1)
		    OR EXISTS (SELECT 1 FROM vendors WHERE business_slug = $1)
	`, slug).Scan(&taken)
	if err != nil {
		return false, apperr.ErrInternal.WithCause(err)
	}
	return taken, nil
}

// HasActiveApplication reports whether the user has a pending/approved application.
func (r *Repository) HasActiveApplication(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM store_applications
			WHERE applicant_user_id = $1 AND status IN ('pending', 'approved')
		)
	`, userID).Scan(&exists)
	if err != nil {
		return false, apperr.ErrInternal.WithCause(err)
	}
	return exists, nil
}

// ApplicationsForUser returns the user's applications, newest first.
func (r *Repository) ApplicationsForUser(ctx context.Context, userID string) ([]Application, error) {
	rows, err := r.db.Query(ctx, `SELECT`+applicationColumns+applicationFrom+`
		WHERE a.applicant_user_id = $1
		ORDER BY a.submitted_at DESC, a.id`, userID)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	return collectApplications(rows)
}

// ApplicationByID loads one application (with joined display fields).
func (r *Repository) ApplicationByID(ctx context.Context, id string) (Application, error) {
	a, err := scanApplication(r.db.QueryRow(ctx, `SELECT`+applicationColumns+applicationFrom+` WHERE a.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Application{}, errNotFound
	}
	if err != nil {
		return Application{}, apperr.ErrInternal.WithCause(err)
	}
	return a, nil
}

// ApplicationPaymentStatus returns the latest payment status for an application
// (empty when there is none).
func (r *Repository) ApplicationPaymentStatus(ctx context.Context, applicationID string) (string, error) {
	var status string
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE((SELECT status::text FROM application_payments
			WHERE application_id = $1 ORDER BY created_at DESC LIMIT 1), '')
	`, applicationID).Scan(&status)
	if err != nil {
		return "", apperr.ErrInternal.WithCause(err)
	}
	return status, nil
}

// ListApplications returns one page of applications for the admin queue.
func (r *Repository) ListApplications(ctx context.Context, q AdminListQuery) ([]Application, int64, error) {
	where, args := buildApplicationFilter(q)

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT count(*)`+applicationFrom+` WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []Application{}, 0, nil
	}

	limitPlaceholder := "$" + strconv.Itoa(len(args)+1)
	offsetPlaceholder := "$" + strconv.Itoa(len(args)+2)
	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)

	order := "a.submitted_at DESC, a.id"
	if q.Sort == "submitted_at_asc" {
		order = "a.submitted_at ASC, a.id"
	}

	rows, err := r.db.Query(ctx,
		`SELECT`+applicationColumns+applicationFrom+` WHERE `+where+`
		ORDER BY `+order+`
		LIMIT `+limitPlaceholder+` OFFSET `+offsetPlaceholder, listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records, err := collectApplications(rows)
	if err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func buildApplicationFilter(q AdminListQuery) (string, []any) {
	conditions := []string{"TRUE"}
	args := []any{}

	if q.Status != "" {
		args = append(args, q.Status)
		conditions = append(conditions, "a.status = $"+strconv.Itoa(len(args))+"::application_status")
	}
	if q.PaymentStatus != "" {
		args = append(args, q.PaymentStatus)
		conditions = append(conditions, "EXISTS (SELECT 1 FROM application_payments ap WHERE ap.application_id = a.id AND ap.status = $"+strconv.Itoa(len(args))+"::payment_status)")
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		place := "$" + strconv.Itoa(len(args))
		conditions = append(conditions, fmt.Sprintf(
			"(a.application_number ILIKE %[1]s OR a.store_name ILIKE %[1]s OR u.email ILIKE %[1]s OR u.id::text ILIKE %[1]s OR p.first_name ILIKE %[1]s OR p.last_name ILIKE %[1]s)",
			place,
		))
	}

	return strings.Join(conditions, " AND "), args
}

// AuditForApplication returns the audit history for an application.
func (r *Repository) AuditForApplication(ctx context.Context, id string) ([]AuditEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT al.action,
		       COALESCE(NULLIF(TRIM(CONCAT_WS(' ', p.first_name, p.last_name)), ''), split_part(au.email, '@', 1), 'System'),
		       al.created_at
		FROM audit_logs al
		LEFT JOIN users au ON au.id = al.user_id
		LEFT JOIN user_profiles p ON p.user_id = al.user_id
		WHERE al.entity_type = 'store_application' AND al.entity_id = $1
		ORDER BY al.created_at ASC`, id)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	entries := make([]AuditEntry, 0)
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.Action, &e.ActorName, &e.CreatedAt); err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return entries, nil
}

// Approve marks an application approved, creates the store + vendor + owner
// membership, and (when no fee is required) activates immediately. All steps
// run in one transaction; the application row is locked to prevent concurrent
// approve/reject.
func (r *Repository) Approve(ctx context.Context, id, adminID string, feeEnabled bool) (Application, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Application{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		status       string
		applicantID  string
		storeName    string
		storeSlug    string
		description  string
		categoryID   *string
		logoURL      *string
		bannerURL    *string
		businessMail string
		contactPhone string
	)
	err = tx.QueryRow(ctx, `
		SELECT status::text, applicant_user_id, store_name, store_slug, description,
		       category_id, logo_url, banner_url, business_email, contact_phone
		FROM store_applications WHERE id = $1 FOR UPDATE`, id,
	).Scan(&status, &applicantID, &storeName, &storeSlug, &description,
		&categoryID, &logoURL, &bannerURL, &businessMail, &contactPhone)
	if errors.Is(err, pgx.ErrNoRows) {
		return Application{}, errNotFound
	}
	if err != nil {
		return Application{}, apperr.ErrInternal.WithCause(err)
	}
	if status != StatusPending {
		return Application{}, errNotReviewable
	}

	if _, err := tx.Exec(ctx, `
		UPDATE store_applications
		SET status = 'approved', reviewed_by = $2, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $1`, id, adminID); err != nil {
		return Application{}, apperr.ErrInternal.WithCause(err)
	}

	var vendorID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO vendors (user_id, business_name, business_slug, business_email, business_phone, description, logo_url, banner_url, category_id, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'pending')
		RETURNING id`,
		applicantID, storeName, storeSlug, businessMail, contactPhone, description, logoURL, bannerURL, categoryID,
	).Scan(&vendorID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return Application{}, errSlugTaken.WithCause(err)
		}
		return Application{}, apperr.ErrInternal.WithCause(err)
	}

	var storeID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO stores (vendor_id, application_id, name, slug, description, category_id, logo_url, banner_url, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending_activation')
		RETURNING id`,
		vendorID, id, storeName, storeSlug, description, categoryID, logoURL, bannerURL,
	).Scan(&storeID); err != nil {
		return Application{}, apperr.ErrInternal.WithCause(err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO store_memberships (store_id, user_id, role, status)
		VALUES ($1, $2, 'owner', 'active')`, storeID, applicantID); err != nil {
		return Application{}, apperr.ErrInternal.WithCause(err)
	}

	if !feeEnabled {
		if err := activateTx(ctx, tx, id); err != nil {
			return Application{}, err
		}
	}

	_ = insertAuditTx(ctx, tx, &adminID, "store_application.approved", id)
	_ = insertNotificationTx(ctx, tx, applicantID, "vendor_approved",
		"Store application approved",
		fmt.Sprintf("Your store \"%s\" was approved. Complete the application fee to activate it.", storeName))

	if err := tx.Commit(ctx); err != nil {
		return Application{}, apperr.ErrInternal.WithCause(err)
	}
	return r.ApplicationByID(ctx, id)
}

// Reject marks a pending application rejected with a reason.
func (r *Repository) Reject(ctx context.Context, id, adminID, reason string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status, applicantID, storeName string
	err = tx.QueryRow(ctx, `
		SELECT status::text, applicant_user_id, store_name
		FROM store_applications WHERE id = $1 FOR UPDATE`, id,
	).Scan(&status, &applicantID, &storeName)
	if errors.Is(err, pgx.ErrNoRows) {
		return errNotFound
	}
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if status != StatusPending {
		return errNotReviewable
	}

	if _, err := tx.Exec(ctx, `
		UPDATE store_applications
		SET status = 'rejected', rejection_reason = $2, reviewed_by = $3, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $1`, id, reason, adminID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	_ = insertAuditTx(ctx, tx, &adminID, "store_application.rejected", id)
	_ = insertNotificationTx(ctx, tx, applicantID, "vendor_rejected",
		"Store application declined",
		fmt.Sprintf("Your store application for \"%s\" was not approved.", storeName))

	return tx.Commit(ctx)
}

// UpdateNotes stores an internal admin note (not exposed to the applicant).
func (r *Repository) UpdateNotes(ctx context.Context, id, note string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE store_applications SET admin_review_notes = $2, updated_at = NOW() WHERE id = $1`, id, note)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return errNotFound
	}
	return nil
}

// StoresForUser lists stores the user is an active member of.
func (r *Repository) StoresForUser(ctx context.Context, userID string) ([]Store, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.id, s.vendor_id, s.application_id, s.name, s.slug, COALESCE(s.description, ''),
		       COALESCE(c.name, ''), COALESCE(s.logo_url, ''), COALESCE(s.banner_url, ''),
		       s.status::text, s.activated_at, s.created_at, m.role::text
		FROM store_memberships m
		JOIN stores s ON s.id = m.store_id
		LEFT JOIN categories c ON c.id = s.category_id
		WHERE m.user_id = $1 AND m.status = 'active'
		ORDER BY s.created_at DESC, s.id`, userID)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	stores := make([]Store, 0)
	for rows.Next() {
		var s Store
		if err := rows.Scan(&s.ID, &s.VendorID, &s.ApplicationID, &s.Name, &s.Slug, &s.Description,
			&s.Category, &s.LogoURL, &s.BannerURL, &s.Status, &s.ActivatedAt, &s.CreatedAt, &s.Role); err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		stores = append(stores, s)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return stores, nil
}

// ResolveStore resolves the store scope for a vendor request: a specific store
// when storeID is given, otherwise the caller's active store. Inactive stores
// are rejected so suspended/pending stores cannot operate.
func (r *Repository) ResolveStore(ctx context.Context, userID, storeID string) (StoreContext, error) {
	var sc StoreContext
	err := r.db.QueryRow(ctx, `
		SELECT s.id, s.vendor_id, s.name, s.status::text, m.role::text
		FROM store_memberships m
		JOIN stores s ON s.id = m.store_id
		WHERE m.user_id = $1 AND m.status = 'active'
		  AND ($2 = '' OR s.id::text = $2)
		ORDER BY (s.status = 'active') DESC, s.created_at DESC
		LIMIT 1`, userID, storeID,
	).Scan(&sc.StoreID, &sc.VendorID, &sc.Name, &sc.Status, &sc.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return StoreContext{}, errNoActiveStore
	}
	if err != nil {
		return StoreContext{}, apperr.ErrInternal.WithCause(err)
	}
	if sc.Status != StoreActive {
		return StoreContext{}, errStoreNotActive
	}
	return sc, nil
}

// StoreForApplication returns the store created for an application, if any.
func (r *Repository) StoreForApplication(ctx context.Context, applicationID string) (*Store, error) {
	var s Store
	err := r.db.QueryRow(ctx, `
		SELECT s.id, s.vendor_id, s.application_id, s.name, s.slug, COALESCE(s.description, ''),
		       COALESCE(c.name, ''), COALESCE(s.logo_url, ''), COALESCE(s.banner_url, ''),
		       s.status::text, s.activated_at, s.created_at
		FROM stores s
		LEFT JOIN categories c ON c.id = s.category_id
		WHERE s.application_id = $1`, applicationID,
	).Scan(&s.ID, &s.VendorID, &s.ApplicationID, &s.Name, &s.Slug, &s.Description,
		&s.Category, &s.LogoURL, &s.BannerURL, &s.Status, &s.ActivatedAt, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return &s, nil
}

// ListStores returns one page of stores for the admin stores list.
func (r *Repository) ListStores(ctx context.Context, q AdminListQuery) ([]Store, int64, error) {
	conditions := []string{"TRUE"}
	args := []any{}
	if q.Status != "" {
		args = append(args, q.Status)
		conditions = append(conditions, "s.status = $"+strconv.Itoa(len(args))+"::store_status")
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		place := "$" + strconv.Itoa(len(args))
		conditions = append(conditions, fmt.Sprintf("(s.name ILIKE %[1]s OR s.slug ILIKE %[1]s)", place))
	}
	where := strings.Join(conditions, " AND ")

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM stores s WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []Store{}, 0, nil
	}

	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.db.Query(ctx, `
		SELECT s.id, s.vendor_id, s.application_id, s.name, s.slug, COALESCE(s.description, ''),
		       COALESCE(c.name, ''), COALESCE(s.logo_url, ''), COALESCE(s.banner_url, ''),
		       s.status::text, s.activated_at, s.created_at
		FROM stores s
		LEFT JOIN categories c ON c.id = s.category_id
		WHERE `+where+`
		ORDER BY s.created_at DESC, s.id
		LIMIT $`+strconv.Itoa(len(args)+1)+` OFFSET $`+strconv.Itoa(len(args)+2), listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	stores := make([]Store, 0)
	for rows.Next() {
		var s Store
		if err := rows.Scan(&s.ID, &s.VendorID, &s.ApplicationID, &s.Name, &s.Slug, &s.Description,
			&s.Category, &s.LogoURL, &s.BannerURL, &s.Status, &s.ActivatedAt, &s.CreatedAt); err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		stores = append(stores, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	return stores, total, nil
}

// CreatePayment inserts a pending application-fee payment.
func (r *Repository) CreatePayment(ctx context.Context, p PaymentParams) (Payment, error) {
	var payment Payment
	var txnID *string
	err := r.db.QueryRow(ctx, `
		INSERT INTO application_payments (application_id, user_id, amount, currency, provider, payment_reference, status)
		VALUES ($1, $2, $3::numeric, $4, $5, $6, 'pending')
		RETURNING id, application_id, user_id, amount::text, currency, COALESCE(provider, ''),
		          payment_reference, provider_transaction_id, status::text, verified_at, created_at`,
		p.ApplicationID, p.UserID, p.Amount, p.Currency, nullable(p.Provider), p.Reference,
	).Scan(&payment.ID, &payment.ApplicationID, &payment.UserID, &payment.Amount, &payment.Currency,
		&payment.Provider, &payment.Reference, &txnID, &payment.Status, &payment.VerifiedAt, &payment.CreatedAt)
	if err != nil {
		return Payment{}, apperr.ErrInternal.WithCause(err)
	}
	payment.ProviderTransactionID = txnID
	return payment, nil
}

// PaymentsForApplication returns all payments for an application, newest first.
func (r *Repository) PaymentsForApplication(ctx context.Context, applicationID string) ([]Payment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, application_id, user_id, amount::text, currency, COALESCE(provider, ''),
		       payment_reference, provider_transaction_id, status::text, verified_at, created_at
		FROM application_payments
		WHERE application_id = $1
		ORDER BY created_at DESC, id`, applicationID)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	payments := make([]Payment, 0)
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.ApplicationID, &p.UserID, &p.Amount, &p.Currency, &p.Provider,
			&p.Reference, &p.ProviderTransactionID, &p.Status, &p.VerifiedAt, &p.CreatedAt); err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		payments = append(payments, p)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return payments, nil
}

// LatestPayment returns the most recent payment for an application, if any.
func (r *Repository) LatestPayment(ctx context.Context, applicationID string) (*Payment, error) {
	var p Payment
	err := r.db.QueryRow(ctx, `
		SELECT id, application_id, user_id, amount::text, currency, COALESCE(provider, ''),
		       payment_reference, provider_transaction_id, status::text, verified_at, created_at
		FROM application_payments
		WHERE application_id = $1
		ORDER BY created_at DESC, id
		LIMIT 1`, applicationID,
	).Scan(&p.ID, &p.ApplicationID, &p.UserID, &p.Amount, &p.Currency, &p.Provider,
		&p.Reference, &p.ProviderTransactionID, &p.Status, &p.VerifiedAt, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return &p, nil
}

// PaymentByReference loads a payment by our reference.
func (r *Repository) PaymentByReference(ctx context.Context, ref string) (Payment, error) {
	var p Payment
	err := r.db.QueryRow(ctx, `
		SELECT id, application_id, user_id, amount::text, currency, COALESCE(provider, ''),
		       payment_reference, provider_transaction_id, status::text, verified_at, created_at
		FROM application_payments WHERE payment_reference = $1`, ref,
	).Scan(&p.ID, &p.ApplicationID, &p.UserID, &p.Amount, &p.Currency, &p.Provider,
		&p.Reference, &p.ProviderTransactionID, &p.Status, &p.VerifiedAt, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, errNotFound
	}
	if err != nil {
		return Payment{}, apperr.ErrInternal.WithCause(err)
	}
	return p, nil
}

// MarkPaid transitions a pending/processing payment to paid. It is guarded so a
// repeated webhook cannot re-transition; the second bool reports whether a
// transition happened.
func (r *Repository) MarkPaid(ctx context.Context, ref, providerTxnID string) (Payment, bool, error) {
	var p Payment
	err := r.db.QueryRow(ctx, `
		UPDATE application_payments
		SET status = 'paid', verified_at = NOW(),
		    provider_transaction_id = COALESCE(NULLIF($2, ''), provider_transaction_id),
		    updated_at = NOW()
		WHERE payment_reference = $1 AND status IN ('pending', 'processing')
		RETURNING id, application_id, user_id, amount::text, currency, COALESCE(provider, ''),
		          payment_reference, provider_transaction_id, status::text, verified_at, created_at`,
		ref, providerTxnID,
	).Scan(&p.ID, &p.ApplicationID, &p.UserID, &p.Amount, &p.Currency, &p.Provider,
		&p.Reference, &p.ProviderTransactionID, &p.Status, &p.VerifiedAt, &p.CreatedAt)
	if err == nil {
		return p, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			// Provider transaction id already recorded for another payment:
			// treat as an idempotent replay rather than an error.
			existing, getErr := r.PaymentByReference(ctx, ref)
			return existing, false, getErr
		}
		return Payment{}, false, apperr.ErrInternal.WithCause(err)
	}
	// No row updated: either already paid or an unknown reference.
	existing, getErr := r.PaymentByReference(ctx, ref)
	if getErr != nil {
		return Payment{}, false, getErr
	}
	return existing, false, nil
}

// ActivateForApplication activates the store for an approved application
// exactly once (idempotent) and grants the applicant the vendor role.
func (r *Repository) ActivateForApplication(ctx context.Context, applicationID string) (Store, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Store{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var applicantID, storeName string
	err = tx.QueryRow(ctx, `
		SELECT status::text, applicant_user_id, store_name
		FROM store_applications WHERE id = $1 FOR UPDATE`, applicationID,
	).Scan(new(string), &applicantID, &storeName)
	if errors.Is(err, pgx.ErrNoRows) {
		return Store{}, errNotFound
	}
	if err != nil {
		return Store{}, apperr.ErrInternal.WithCause(err)
	}

	if err := activateTx(ctx, tx, applicationID); err != nil {
		return Store{}, err
	}

	_ = insertAuditTx(ctx, tx, nil, "store.activated", applicationID)
	_ = insertNotificationTx(ctx, tx, applicantID, "payment_success",
		"Your store is live",
		fmt.Sprintf("Your store \"%s\" is active. Head to your vendor dashboard to start selling.", storeName))

	if err := tx.Commit(ctx); err != nil {
		return Store{}, apperr.ErrInternal.WithCause(err)
	}

	store, err := r.StoreForApplication(ctx, applicationID)
	if err != nil {
		return Store{}, err
	}
	if store == nil {
		return Store{}, errNotActivated
	}
	return *store, nil
}

// activateTx flips the store to active, the vendor to approved, and grants the
// applicant the vendor role. Idempotent: an already-active store is a no-op.
func activateTx(ctx context.Context, tx pgx.Tx, applicationID string) error {
	var (
		storeID     string
		vendorID    string
		status      string
		applicantID string
	)
	err := tx.QueryRow(ctx, `
		SELECT s.id, s.vendor_id, s.status::text, a.applicant_user_id
		FROM stores s
		JOIN store_applications a ON a.id = s.application_id
		WHERE s.application_id = $1
		FOR UPDATE`, applicationID,
	).Scan(&storeID, &vendorID, &status, &applicantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return errNotActivated
	}
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if status == StoreActive {
		return nil
	}

	if _, err := tx.Exec(ctx, `
		UPDATE stores SET status = 'active', activated_at = NOW(), updated_at = NOW() WHERE id = $1`, storeID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE vendors SET status = 'approved', approved_at = NOW(), updated_at = NOW() WHERE id = $1`, vendorID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = 'vendor'::user_role_type
		ON CONFLICT DO NOTHING`, applicantID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// ListPayments returns one page of application payments for the admin list.
func (r *Repository) ListPayments(ctx context.Context, q AdminListQuery) ([]PaymentAdminRow, int64, error) {
	conditions := []string{"TRUE"}
	args := []any{}
	if q.Status != "" {
		args = append(args, q.Status)
		conditions = append(conditions, "ap.status = $"+strconv.Itoa(len(args))+"::payment_status")
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		place := "$" + strconv.Itoa(len(args))
		conditions = append(conditions, fmt.Sprintf(
			"(ap.payment_reference ILIKE %[1]s OR a.application_number ILIKE %[1]s OR u.email ILIKE %[1]s)", place))
	}
	where := strings.Join(conditions, " AND ")

	const from = `
		FROM application_payments ap
		JOIN store_applications a ON a.id = ap.application_id
		JOIN users u ON u.id = ap.user_id
		LEFT JOIN user_profiles p ON p.user_id = u.id`

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT count(*)`+from+` WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []PaymentAdminRow{}, 0, nil
	}

	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.db.Query(ctx, `
		SELECT ap.id, ap.application_id, ap.user_id, ap.amount::text, ap.currency, COALESCE(ap.provider, ''),
		       ap.payment_reference, ap.provider_transaction_id, ap.status::text, ap.verified_at, ap.created_at,
		       a.application_number, a.store_name,
		       COALESCE(NULLIF(TRIM(CONCAT_WS(' ', p.first_name, p.last_name)), ''), split_part(u.email, '@', 1)) AS applicant_name,
		       u.email`+from+`
		WHERE `+where+`
		ORDER BY ap.created_at DESC, ap.id
		LIMIT $`+strconv.Itoa(len(args)+1)+` OFFSET $`+strconv.Itoa(len(args)+2), listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	out := make([]PaymentAdminRow, 0)
	for rows.Next() {
		var row PaymentAdminRow
		if err := rows.Scan(
			&row.ID, &row.ApplicationID, &row.UserID, &row.Amount, &row.Currency, &row.Provider,
			&row.Reference, &row.ProviderTransactionID, &row.Status, &row.VerifiedAt, &row.CreatedAt,
			&row.ApplicationNumber, &row.StoreName, &row.ApplicantName, &row.ApplicantEmail,
		); err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	return out, total, nil
}

func collectApplications(rows pgx.Rows) ([]Application, error) {
	records := make([]Application, 0)
	for rows.Next() {
		a, err := scanApplication(rows)
		if err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, a)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return records, nil
}

func insertAuditTx(ctx context.Context, tx pgx.Tx, actorID *string, action, entityID string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id)
		VALUES ($1, $2, 'store_application', $3)`, actorID, action, entityID)
	return err
}

func insertNotificationTx(ctx context.Context, tx pgx.Tx, userID, ntype, title, message string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO notifications (user_id, type, title, message)
		VALUES ($1, $2::notification_type, $3, $4)`, userID, ntype, title, message)
	return err
}
