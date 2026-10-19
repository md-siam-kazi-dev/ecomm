package cart

import "context"

// fakeRepo is a configurable in-memory CartRepository for tests.
type fakeRepo struct {
	userCart    Cart
	hasUserCart bool

	sessionCart    Cart
	hasSessionCart bool

	createErr error

	mergedFrom string
	mergedInto string

	variant    VariantResolution
	variantErr error

	items   []ItemRecord
	listErr error

	upsertErr   error
	lastUpsert  upsertArgs
	setQtyErr   error
	lastSetQty  setQtyArgs
	removeErr   error
	removedItem string
	clearErr    error
	markedWith  []string
}

type upsertArgs struct {
	cartID, variantID string
	quantity          int
	unitPrice         float64
	maxQuantity       int
}

type setQtyArgs struct {
	cartID, itemID string
	quantity       int
}

func (f *fakeRepo) FindUserCart(context.Context, string) (Cart, bool, error) {
	return f.userCart, f.hasUserCart, nil
}

func (f *fakeRepo) FindSessionCart(context.Context, string) (Cart, bool, error) {
	return f.sessionCart, f.hasSessionCart, nil
}

func (f *fakeRepo) CreateCart(_ context.Context, userID, sessionID, currency string) (Cart, error) {
	if f.createErr != nil {
		return Cart{}, f.createErr
	}
	return Cart{ID: "cart-new", UserID: userID, SessionID: sessionID, Currency: currency}, nil
}

func (f *fakeRepo) MergeCarts(_ context.Context, from, into string) error {
	f.mergedFrom, f.mergedInto = from, into
	return nil
}

func (f *fakeRepo) MarkCartStatus(_ context.Context, _ string, status string) error {
	f.markedWith = append(f.markedWith, status)
	return nil
}

func (f *fakeRepo) ResolveVariant(context.Context, string, string) (VariantResolution, error) {
	return f.variant, f.variantErr
}

func (f *fakeRepo) ListItems(context.Context, string) ([]ItemRecord, error) {
	return f.items, f.listErr
}

func (f *fakeRepo) UpsertItem(_ context.Context, cartID, variantID string, quantity int, unitPrice float64, maxQuantity int) error {
	f.lastUpsert = upsertArgs{cartID, variantID, quantity, unitPrice, maxQuantity}
	return f.upsertErr
}

func (f *fakeRepo) FindItem(context.Context, string, string) (ItemRecord, error) {
	if f.setQtyErr != nil {
		return ItemRecord{}, f.setQtyErr
	}
	return f.items[0], nil
}

func (f *fakeRepo) SetItemQuantity(_ context.Context, cartID, itemID string, quantity int) error {
	f.lastSetQty = setQtyArgs{cartID, itemID, quantity}
	return f.setQtyErr
}

func (f *fakeRepo) RemoveItem(_ context.Context, _, itemID string) error {
	f.removedItem = itemID
	return f.removeErr
}

func (f *fakeRepo) ClearItems(context.Context, string) error {
	return f.clearErr
}
