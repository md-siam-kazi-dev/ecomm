package order

import (
	"context"
	"errors"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/payment"
)

func TestCreate_RejectsEmptyItems(t *testing.T) {
	req := validRequest(MethodCOD)
	req.Items = nil
	_, err := NewService(&fakeRepo{}, payment.NewRegistry(), testConfig("")).Create(context.Background(), "u-1", req)
	if !errors.Is(err, errEmptyItems) {
		t.Fatalf("expected errEmptyItems, got %v", err)
	}
}

func TestCreate_RejectsBadPaymentMethod(t *testing.T) {
	req := validRequest("crypto")
	_, err := NewService(&fakeRepo{}, payment.NewRegistry(), testConfig("")).Create(context.Background(), "u-1", req)
	if !errors.Is(err, errBadPayment) {
		t.Fatalf("expected errBadPayment, got %v", err)
	}
}

func TestCreate_ComputesTotalsServerSide(t *testing.T) {
	repo := &fakeRepo{resolved: []VariantResolution{sampleResolution()}}
	svc := NewService(repo, payment.NewRegistry(), testConfig(""))

	out, err := svc.Create(context.Background(), "u-1", validRequest(MethodCOD))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastParams.Subtotal != 39.99 {
		t.Errorf("subtotal = %v, want 39.99", repo.lastParams.Subtotal)
	}
	if repo.lastParams.ShippingTotal != 5.99 {
		t.Errorf("shipping = %v, want 5.99", repo.lastParams.ShippingTotal)
	}
	if repo.lastParams.GrandTotal != 45.98 {
		t.Errorf("grand total = %v, want 45.98", repo.lastParams.GrandTotal)
	}
	if out.CheckoutURL != "" {
		t.Errorf("cod should not return a checkout url, got %q", out.CheckoutURL)
	}
	if repo.paymentCreated {
		t.Error("cod must not create a payment")
	}
}

func TestCreate_FreeShippingOverThreshold(t *testing.T) {
	line := sampleResolution()
	line.UnitPrice = 100
	repo := &fakeRepo{resolved: []VariantResolution{line}}
	svc := NewService(repo, payment.NewRegistry(), testConfig(""))

	if _, err := svc.Create(context.Background(), "u-1", validRequest(MethodCOD)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastParams.ShippingTotal != 0 {
		t.Errorf("shipping = %v, want 0 (free over threshold)", repo.lastParams.ShippingTotal)
	}
}

func TestCreate_InsufficientStock(t *testing.T) {
	line := sampleResolution()
	line.Available = 1
	repo := &fakeRepo{resolved: []VariantResolution{line}}
	req := validRequest(MethodCOD)
	req.Items[0].Quantity = 2

	_, err := NewService(repo, payment.NewRegistry(), testConfig("")).Create(context.Background(), "u-1", req)
	if !errors.Is(err, errInsufficientStock) {
		t.Fatalf("expected errInsufficientStock, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Error("order must not be created when stock is insufficient")
	}
}

func TestCreate_OnlineRequiresProvider(t *testing.T) {
	repo := &fakeRepo{resolved: []VariantResolution{sampleResolution()}}
	// Registry has no "stripe" provider registered.
	_, err := NewService(repo, payment.NewRegistry(), testConfig("stripe")).Create(context.Background(), "u-1", validRequest(MethodCard))
	if !errors.Is(err, errPaymentNotConfigured) {
		t.Fatalf("expected errPaymentNotConfigured, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Error("order must not be created when online payment is not configured")
	}
}

func TestCreate_OnlineReturnsCheckoutURL(t *testing.T) {
	repo := &fakeRepo{resolved: []VariantResolution{sampleResolution()}}
	provider := &fakeProvider{name: "fake", redirect: "https://pay.test/session"}
	svc := NewService(repo, registryWith(provider), testConfig("fake"))

	out, err := svc.Create(context.Background(), "u-1", validRequest(MethodCard))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CheckoutURL != "https://pay.test/session" {
		t.Errorf("checkout url = %q", out.CheckoutURL)
	}
	if !repo.paymentCreated {
		t.Error("online payment should create a payment row")
	}
}

func TestCreate_OnlineInitFailureCompensates(t *testing.T) {
	repo := &fakeRepo{resolved: []VariantResolution{sampleResolution()}}
	provider := &fakeProvider{name: "fake", initErr: errors.New("gateway down")}
	svc := NewService(repo, registryWith(provider), testConfig("fake"))

	if _, err := svc.Create(context.Background(), "u-1", validRequest(MethodCard)); err == nil {
		t.Fatal("expected an error when the provider fails")
	}
	if repo.deletedOrderID == "" {
		t.Error("a failed payment session must roll the order back (no orphan order)")
	}
}

func TestCreate_PaymentRowFailureCompensates(t *testing.T) {
	repo := &fakeRepo{resolved: []VariantResolution{sampleResolution()}, paymentErr: errors.New("db down")}
	provider := &fakeProvider{name: "fake", redirect: "https://pay.test/session"}
	svc := NewService(repo, registryWith(provider), testConfig("fake"))

	if _, err := svc.Create(context.Background(), "u-1", validRequest(MethodCard)); err == nil {
		t.Fatal("expected an error when the payment row cannot be created")
	}
	if repo.deletedOrderID == "" {
		t.Error("a failed payment row must roll the order back")
	}
}

func TestCreate_RetriesOrderNumberCollision(t *testing.T) {
	repo := &fakeRepo{resolved: []VariantResolution{sampleResolution()}}
	svc := NewService(&collisionRepo{fakeRepo: repo, remaining: 2}, payment.NewRegistry(), testConfig(""))

	if _, err := svc.Create(context.Background(), "u-1", validRequest(MethodCOD)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createCalls < 3 {
		t.Errorf("expected retries, createCalls=%d", repo.createCalls)
	}
}

// collisionRepo returns errOrderNumberTaken the first N times.
type collisionRepo struct {
	*fakeRepo
	remaining int
}

func (c *collisionRepo) CreateOrder(ctx context.Context, p OrderParams) (Order, error) {
	c.createCalls++
	c.lastParams = p
	if c.remaining > 0 {
		c.remaining--
		return Order{}, errOrderNumberTaken
	}
	return Order{ID: "order-1", OrderNumber: p.OrderNumber, Status: StatusPending}, nil
}

func TestHandleWebhook_UnpaidNoop(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, registryWith(&fakeProvider{name: "fake", paid: false}), testConfig("fake"))

	if err := svc.HandleWebhook(context.Background(), "fake", payment.VerifyRequest{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.confirmCalls != 0 {
		t.Error("unpaid event must not confirm the order")
	}
}

func TestHandleWebhook_VerifiedConfirms(t *testing.T) {
	repo := &fakeRepo{paymentByRef: Payment{ID: "pay-1", Amount: 45.98, Currency: "usd"}}
	svc := NewService(repo, registryWith(&fakeProvider{name: "fake", paid: true}), testConfig("fake"))

	if err := svc.HandleWebhook(context.Background(), "fake", payment.VerifyRequest{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.confirmCalls != 1 {
		t.Errorf("confirmCalls = %d, want 1", repo.confirmCalls)
	}
	if !repo.txnInserted {
		t.Error("expected a payment transaction row")
	}
}

func TestGet_NotFound(t *testing.T) {
	repo := &fakeRepo{orderErr: errNotFound}
	_, err := NewService(repo, payment.NewRegistry(), testConfig("")).Get(context.Background(), "u-1", "missing")
	if !errors.Is(err, errNotFound) {
		t.Fatalf("expected errNotFound, got %v", err)
	}
}

func TestList_NormalizesLimit(t *testing.T) {
	repo := &fakeRepo{}
	_, pagination, err := NewService(repo, payment.NewRegistry(), testConfig("")).List(context.Background(), "u-1", 0, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pagination.Page != 1 || pagination.Limit != defaultLimit {
		t.Errorf("pagination not normalized: %+v", pagination)
	}
}
