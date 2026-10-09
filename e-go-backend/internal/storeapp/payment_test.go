package storeapp

import (
	"context"
	"errors"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/payment"
)

func feeConfig() Config {
	return Config{
		FeeEnabled:   true,
		FeeAmount:    "1500.00",
		FeeCurrency:  "BDT",
		ProviderName: "fake",
		PlatformName: "ShopKart",
	}
}

func TestInitiatePayment_FeeDisabled(t *testing.T) {
	repo := newFakeRepo()
	repo.app = approvedApp()
	_, err := NewService(repo, nil, Config{FeeEnabled: false}).InitiatePayment(context.Background(), "u-1", "app-1")
	if !errors.Is(err, errPaymentNotRequired) {
		t.Fatalf("expected errPaymentNotRequired, got %v", err)
	}
}

func TestInitiatePayment_NotConfigured(t *testing.T) {
	repo := newFakeRepo()
	repo.app = approvedApp()
	// ProviderName empty -> NotConfigured placeholder.
	svc := NewService(repo, payment.NewRegistry(), Config{
		FeeEnabled: true, FeeAmount: "1500.00", FeeCurrency: "BDT",
	})

	_, err := svc.InitiatePayment(context.Background(), "u-1", "app-1")
	if !errors.Is(err, errPaymentNotConfigured) {
		t.Fatalf("expected errPaymentNotConfigured, got %v", err)
	}
	if repo.lastPaymentParams.Reference != "" {
		t.Error("no payment record should be created when the provider is not configured")
	}
}

func TestInitiatePayment_Configured(t *testing.T) {
	repo := newFakeRepo()
	repo.app = approvedApp()
	repo.createPaymentRes = Payment{Reference: "PAY-1", Amount: "1500.00", Currency: "BDT", Status: PaymentPending}

	svc := NewService(repo, registryWith(&fakeProvider{name: "fake"}), feeConfig())

	out, err := svc.InitiatePayment(context.Background(), "u-1", "app-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.ProviderConfigured || out.RedirectURL == "" {
		t.Fatalf("expected configured redirect, got %+v", out)
	}
	if repo.lastPaymentParams.Amount != "1500.00" || repo.lastPaymentParams.Currency != "BDT" {
		t.Errorf("fee not taken from server config: %+v", repo.lastPaymentParams)
	}
}

func TestInitiatePayment_NotOwner(t *testing.T) {
	repo := newFakeRepo()
	repo.app = approvedApp() // applicant u-1
	svc := NewService(repo, registryWith(&fakeProvider{name: "fake"}), feeConfig())

	_, err := svc.InitiatePayment(context.Background(), "someone-else", "app-1")
	if !errors.Is(err, errForbidden) {
		t.Fatalf("expected errForbidden, got %v", err)
	}
}

func TestInitiatePayment_NotApproved(t *testing.T) {
	repo := newFakeRepo()
	app := approvedApp()
	app.Status = StatusPending
	repo.app = app
	svc := NewService(repo, registryWith(&fakeProvider{name: "fake"}), feeConfig())

	_, err := svc.InitiatePayment(context.Background(), "u-1", "app-1")
	if !errors.Is(err, errNotPayable) {
		t.Fatalf("expected errNotPayable, got %v", err)
	}
}

func TestWebhook_UnknownProvider(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, payment.NewRegistry(), feeConfig())

	_, err := svc.HandleWebhook(context.Background(), "nope", payment.VerifyRequest{Reference: "PAY-1"})
	if !errors.Is(err, errPaymentNotConfigured) {
		t.Fatalf("expected errPaymentNotConfigured, got %v", err)
	}
}

func TestWebhook_UnverifiedDoesNotActivate(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, registryWith(&fakeProvider{name: "fake", paid: false}), feeConfig())

	_, err := svc.HandleWebhook(context.Background(), "fake", payment.VerifyRequest{Reference: "PAY-1"})
	if err == nil {
		t.Fatal("expected an error for an unverified payment")
	}
	if repo.activateCalls != 0 {
		t.Errorf("activation must not run for an unverified payment (calls=%d)", repo.activateCalls)
	}
}

func TestWebhook_VerifiedActivates(t *testing.T) {
	repo := newFakeRepo()
	repo.paymentByRef = Payment{ApplicationID: "app-1", Reference: "PAY-1", Status: PaymentPending}
	repo.markPaidRes = Payment{ApplicationID: "app-1", Reference: "PAY-1", Status: PaymentPaid}
	repo.markPaidChanged = true

	svc := NewService(repo, registryWith(&fakeProvider{name: "fake", paid: true}), feeConfig())

	item, err := svc.HandleWebhook(context.Background(), "fake", payment.VerifyRequest{Reference: "PAY-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Status != PaymentPaid {
		t.Errorf("expected paid item, got %+v", item)
	}
	if repo.activateCalls != 1 {
		t.Errorf("expected one activation attempt, got %d", repo.activateCalls)
	}
}

func TestWebhook_RepeatedIsIdempotent(t *testing.T) {
	repo := newFakeRepo()
	repo.paymentByRef = Payment{ApplicationID: "app-1", Reference: "PAY-1", Status: PaymentPaid}
	repo.markPaidRes = Payment{ApplicationID: "app-1", Reference: "PAY-1", Status: PaymentPaid}
	repo.markPaidChanged = false // already paid

	svc := NewService(repo, registryWith(&fakeProvider{name: "fake", paid: true}), feeConfig())

	// Two deliveries of the same callback.
	for i := 0; i < 2; i++ {
		if _, err := svc.HandleWebhook(context.Background(), "fake", payment.VerifyRequest{Reference: "PAY-1"}); err != nil {
			t.Fatalf("call %d unexpected error: %v", i, err)
		}
	}
	// Activation is idempotent in the repository (DB-enforced); the service
	// still attempts it safely on each verified callback.
	if repo.activateCalls != 2 {
		t.Errorf("expected 2 safe activation attempts, got %d", repo.activateCalls)
	}
}
