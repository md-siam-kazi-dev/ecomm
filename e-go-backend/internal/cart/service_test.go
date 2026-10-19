package cart

import (
	"context"
	"errors"
	"testing"
)

func TestGet_GuestWithoutSession_CreatesCartAndCookie(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, "usd")

	out, res, err := svc.Get(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NewSession == "" {
		t.Fatal("expected a new session id for a first-time guest")
	}
	if res.ClearSession {
		t.Fatal("must not clear a cookie it just created")
	}
	if out.Currency != "usd" || out.ItemCount != 0 || len(out.Items) != 0 {
		t.Fatalf("unexpected cart: %+v", out)
	}
}

func TestGet_GuestWithSession_ReusesCart(t *testing.T) {
	repo := &fakeRepo{sessionCart: Cart{ID: "c-1", SessionID: "s-1", Currency: "usd"}, hasSessionCart: true}
	svc := NewService(repo, "usd")

	_, res, err := svc.Get(context.Background(), "", "s-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NewSession != "" {
		t.Fatalf("must not mint a new session when one exists, got %q", res.NewSession)
	}
}

func TestGet_UserCartMergesGuestAndClearsCookie(t *testing.T) {
	repo := &fakeRepo{
		userCart:       Cart{ID: "u-cart", UserID: "u-1", Currency: "usd"},
		hasUserCart:    true,
		sessionCart:    Cart{ID: "g-cart", SessionID: "s-1", Currency: "usd"},
		hasSessionCart: true,
	}
	svc := NewService(repo, "usd")

	_, res, err := svc.Get(context.Background(), "u-1", "s-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.mergedFrom != "g-cart" || repo.mergedInto != "u-cart" {
		t.Fatalf("expected guest→user merge, got %q→%q", repo.mergedFrom, repo.mergedInto)
	}
	if !res.ClearSession {
		t.Fatal("expected the guest cookie to be cleared after merge")
	}
}

func TestAdd_ResolvesVariantAndClampsToStock(t *testing.T) {
	repo := &fakeRepo{
		userCart:    Cart{ID: "u-cart", UserID: "u-1", Currency: "usd"},
		hasUserCart: true,
		variant:     VariantResolution{ProductID: "p-1", VariantID: "v-1", UnitPrice: 39.99, Available: 200},
	}
	svc := NewService(repo, "usd")

	_, _, err := svc.Add(context.Background(), "u-1", "", AddItemRequest{ProductID: "p-1", Quantity: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastUpsert.variantID != "v-1" || repo.lastUpsert.quantity != 5 {
		t.Fatalf("unexpected upsert: %+v", repo.lastUpsert)
	}
	if repo.lastUpsert.maxQuantity != MaxLineQuantity {
		t.Fatalf("max quantity = %d, want %d (capped)", repo.lastUpsert.maxQuantity, MaxLineQuantity)
	}
}

func TestAdd_RejectsMissingProduct(t *testing.T) {
	svc := NewService(&fakeRepo{}, "usd")
	_, _, err := svc.Add(context.Background(), "u-1", "", AddItemRequest{ProductID: "  ", Quantity: 1})
	if !errors.Is(err, errProductUnavailable) {
		t.Fatalf("got %v, want errProductUnavailable", err)
	}
}

func TestAdd_RejectsNonPositiveQuantity(t *testing.T) {
	svc := NewService(&fakeRepo{}, "usd")
	_, _, err := svc.Add(context.Background(), "u-1", "", AddItemRequest{ProductID: "p-1", Quantity: 0})
	if !errors.Is(err, errInvalidQuantity) {
		t.Fatalf("got %v, want errInvalidQuantity", err)
	}
}

func TestAdd_RejectsOutOfStock(t *testing.T) {
	repo := &fakeRepo{
		userCart:    Cart{ID: "u-cart", UserID: "u-1", Currency: "usd"},
		hasUserCart: true,
		variant:     VariantResolution{ProductID: "p-1", VariantID: "v-1", UnitPrice: 10, Available: 0},
	}
	svc := NewService(repo, "usd")
	_, _, err := svc.Add(context.Background(), "u-1", "", AddItemRequest{ProductID: "p-1", Quantity: 1})
	if !errors.Is(err, errInsufficientStock) {
		t.Fatalf("got %v, want errInsufficientStock", err)
	}
}

func TestSetQuantity_RejectsAboveStock(t *testing.T) {
	repo := &fakeRepo{
		userCart:    Cart{ID: "u-cart", UserID: "u-1", Currency: "usd"},
		hasUserCart: true,
		items:       []ItemRecord{{ID: "i-1", Available: 3}},
	}
	svc := NewService(repo, "usd")

	qty := 4
	_, _, err := svc.SetQuantity(context.Background(), "u-1", "", "i-1", &qty)
	if !errors.Is(err, errInsufficientStock) {
		t.Fatalf("got %v, want errInsufficientStock", err)
	}
}

func TestSetQuantity_RejectsNilQuantity(t *testing.T) {
	svc := NewService(&fakeRepo{}, "usd")
	_, _, err := svc.SetQuantity(context.Background(), "u-1", "", "i-1", nil)
	if !errors.Is(err, errInvalidBody) {
		t.Fatalf("got %v, want errInvalidBody", err)
	}
}

func TestSetQuantity_RemovesOnZero(t *testing.T) {
	repo := &fakeRepo{
		userCart:    Cart{ID: "u-cart", UserID: "u-1", Currency: "usd"},
		hasUserCart: true,
		items:       []ItemRecord{{ID: "i-1", Available: 3}},
	}
	svc := NewService(repo, "usd")

	qty := 0
	_, _, err := svc.SetQuantity(context.Background(), "u-1", "", "i-1", &qty)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastSetQty.quantity != 0 {
		t.Fatalf("expected quantity 0 passed to repo, got %d", repo.lastSetQty.quantity)
	}
}

func TestToResponse_ComputesTotalsAndMaxQuantity(t *testing.T) {
	c := Cart{ID: "c-1", Currency: "usd"}
	records := []ItemRecord{
		{ID: "i-1", UnitPrice: 10, Quantity: 2, Available: 5},
		{ID: "i-2", UnitPrice: 2.5, Quantity: 3, Available: 0},
	}

	out := toResponse(c, records)
	if out.ItemCount != 5 {
		t.Fatalf("item count = %d, want 5", out.ItemCount)
	}
	if out.Subtotal != 27.5 {
		t.Fatalf("subtotal = %v, want 27.5", out.Subtotal)
	}
	if out.Items[0].MaxQuantity != 5 || !out.Items[0].InStock {
		t.Fatalf("line 0 max/in-stock = %d/%v", out.Items[0].MaxQuantity, out.Items[0].InStock)
	}
	// Stock below the held quantity must not report a max under the quantity.
	if out.Items[1].MaxQuantity != 3 {
		t.Fatalf("line 1 max = %d, want 3", out.Items[1].MaxQuantity)
	}
	if out.Items[1].InStock {
		t.Fatal("line 1 should be out of stock")
	}
}
