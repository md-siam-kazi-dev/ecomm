package storeapp

import (
	"context"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/payment"
)

// fakeRepo is a configurable in-memory StoreRepository for service tests.
type fakeRepo struct {
	created          Application
	createErr        error
	lastCreateParams ApplicationParams

	activeExists bool
	activeErr    error
	slugTaken    bool
	slugErr      error

	apps   []Application
	app    Application
	appErr error

	listApps      []Application
	listAppsTotal int64
	lastListQuery AdminListQuery

	approveErr    error
	approveCalled bool
	rejectErr     error
	lastReject    string
	notesErr      error
	audit         []AuditEntry

	stores      []Store
	storeForApp *Store
	scope       StoreContext
	scopeErr    error

	createPaymentRes  Payment
	createPaymentErr  error
	lastPaymentParams PaymentParams
	payments          []Payment
	latest            *Payment
	paymentByRef      Payment
	paymentByRefErr   error
	markPaidRes       Payment
	markPaidChanged   bool
	markPaidErr       error
	activateCalls     int
	activateErr       error
	listPayments      []PaymentAdminRow
	listStores        []Store
	listTotal         int64
}

func (f *fakeRepo) CreateApplication(_ context.Context, p ApplicationParams) (Application, error) {
	f.lastCreateParams = p
	if f.createErr != nil {
		return Application{}, f.createErr
	}
	return f.created, nil
}

func (f *fakeRepo) SlugTaken(context.Context, string) (bool, error) { return f.slugTaken, f.slugErr }

func (f *fakeRepo) HasActiveApplication(context.Context, string) (bool, error) {
	return f.activeExists, f.activeErr
}

func (f *fakeRepo) ApplicationsForUser(context.Context, string) ([]Application, error) {
	return f.apps, nil
}

func (f *fakeRepo) ApplicationByID(context.Context, string) (Application, error) {
	if f.appErr != nil {
		return Application{}, f.appErr
	}
	return f.app, nil
}

func (f *fakeRepo) ListApplications(_ context.Context, q AdminListQuery) ([]Application, int64, error) {
	f.lastListQuery = q
	return f.listApps, f.listAppsTotal, nil
}

func (f *fakeRepo) ApplicationPaymentStatus(context.Context, string) (string, error) {
	return "", nil
}

func (f *fakeRepo) AuditForApplication(context.Context, string) ([]AuditEntry, error) {
	return f.audit, nil
}

func (f *fakeRepo) Approve(_ context.Context, _, _ string, _ bool) (Application, error) {
	f.approveCalled = true
	if f.approveErr != nil {
		return Application{}, f.approveErr
	}
	return f.app, nil
}

func (f *fakeRepo) Reject(_ context.Context, _, _, reason string) error {
	f.lastReject = reason
	return f.rejectErr
}

func (f *fakeRepo) UpdateNotes(context.Context, string, string) error { return f.notesErr }

func (f *fakeRepo) StoresForUser(context.Context, string) ([]Store, error) { return f.stores, nil }

func (f *fakeRepo) ResolveStore(context.Context, string, string) (StoreContext, error) {
	return f.scope, f.scopeErr
}

func (f *fakeRepo) StoreForApplication(context.Context, string) (*Store, error) {
	return f.storeForApp, nil
}

func (f *fakeRepo) ListStores(_ context.Context, q AdminListQuery) ([]Store, int64, error) {
	f.lastListQuery = q
	return f.listStores, f.listTotal, nil
}

func (f *fakeRepo) CreatePayment(_ context.Context, p PaymentParams) (Payment, error) {
	f.lastPaymentParams = p
	if f.createPaymentErr != nil {
		return Payment{}, f.createPaymentErr
	}
	return f.createPaymentRes, nil
}

func (f *fakeRepo) PaymentsForApplication(context.Context, string) ([]Payment, error) {
	return f.payments, nil
}

func (f *fakeRepo) LatestPayment(context.Context, string) (*Payment, error) { return f.latest, nil }

func (f *fakeRepo) PaymentByReference(context.Context, string) (Payment, error) {
	return f.paymentByRef, f.paymentByRefErr
}

func (f *fakeRepo) MarkPaid(context.Context, string, string) (Payment, bool, error) {
	return f.markPaidRes, f.markPaidChanged, f.markPaidErr
}

func (f *fakeRepo) ActivateForApplication(context.Context, string) (Store, error) {
	f.activateCalls++
	if f.activateErr != nil {
		return Store{}, f.activateErr
	}
	return Store{ID: "s-1", Status: StoreActive}, nil
}

func (f *fakeRepo) ListPayments(_ context.Context, q AdminListQuery) ([]PaymentAdminRow, int64, error) {
	f.lastListQuery = q
	return f.listPayments, f.listTotal, nil
}

// fakeProvider is a configurable payment.Provider for tests.
type fakeProvider struct {
	name      string
	paid      bool
	initErr   error
	verifyErr error
	lastReq   payment.VerifyRequest
}

func (f *fakeProvider) Name() string { return f.name }

func (f *fakeProvider) Initiate(context.Context, payment.InitRequest) (payment.InitResult, error) {
	if f.initErr != nil {
		return payment.InitResult{}, f.initErr
	}
	return payment.InitResult{Provider: f.name, RedirectURL: "https://pay.example/session"}, nil
}

func (f *fakeProvider) Verify(_ context.Context, req payment.VerifyRequest) (payment.VerifyResult, error) {
	f.lastReq = req
	if f.verifyErr != nil {
		return payment.VerifyResult{}, f.verifyErr
	}
	return payment.VerifyResult{
		Reference:             req.Reference,
		ProviderTransactionID: "txn-1",
		Paid:                  f.paid,
	}, nil
}

func registryWith(p payment.Provider) *payment.Registry {
	r := payment.NewRegistry()
	r.Register(p)
	return r
}

func newFakeRepo() *fakeRepo { return &fakeRepo{} }

func approvedApp() Application {
	reviewed := time.Date(2026, 1, 5, 10, 0, 0, 0, time.UTC)
	cat := "cat-1"
	return Application{
		ID:                "app-1",
		ApplicationNumber: "SA-20260105-ABC123",
		ApplicantUserID:   "u-1",
		ApplicantName:     "Customer One",
		ApplicantEmail:    "customer@example.com",
		ApplicantStatus:   "active",
		StoreName:         "Demo Store",
		StoreSlug:         "demo-store",
		Description:       "A demo store.",
		CategoryID:        &cat,
		Category:          "Electronics & Gadgets",
		BusinessEmail:     "biz@example.com",
		ContactPhone:      "+8801000000000",
		Country:           "Bangladesh",
		Division:          "Dhaka",
		District:          "Dhaka",
		FullAddress:       "1 Demo Road",
		BusinessType:      BusinessIndividual,
		Status:            StatusApproved,
		ReviewedAt:        &reviewed,
		SubmittedAt:       time.Date(2026, 1, 4, 9, 0, 0, 0, time.UTC),
		UpdatedAt:         reviewed,
	}
}

func validSubmit() SubmitRequest {
	return SubmitRequest{
		StoreName:     "Demo Store",
		StoreSlug:     "demo-store",
		Description:   "A demo store.",
		CategoryID:    "cat-1",
		BusinessEmail: "biz@example.com",
		ContactPhone:  "+8801000000000",
		Country:       "Bangladesh",
		Division:      "Dhaka",
		District:      "Dhaka",
		FullAddress:   "1 Demo Road",
		BusinessType:  BusinessIndividual,
		AgreedToTerms: true,
	}
}
