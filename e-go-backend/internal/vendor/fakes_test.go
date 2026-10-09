package vendor

import (
	"context"
	"time"
)

type fakeRepo struct {
	scope    StoreScope
	scopeErr error

	summary    Summary
	summaryErr error

	products      []ProductRecord
	productsTotal int64
	orders        []OrderRecord
	ordersTotal   int64

	lastProductQuery ListQuery
	lastOrderQuery   ListQuery
	calls            int

	lastProductParams ProductParams
	lastProductID     string
	createdDetail     ProductDetail
	productErr        error

	inventory          []InventoryRecord
	inventoryTotal     int64
	lastInventoryQuery ListQuery
	lastAdjustQuantity int
	lastAdjustVariant  string
	adjustErr          error
	adjusted           InventoryRecord
}

func (f *fakeRepo) CreateProduct(_ context.Context, _ string, p ProductParams) (ProductDetail, error) {
	f.lastProductParams = p
	if f.productErr != nil {
		return ProductDetail{}, f.productErr
	}
	return f.createdDetail, nil
}

func (f *fakeRepo) UpdateProduct(_ context.Context, _, productID string, p ProductParams) (ProductDetail, error) {
	f.lastProductID, f.lastProductParams = productID, p
	if f.productErr != nil {
		return ProductDetail{}, f.productErr
	}
	return f.createdDetail, nil
}

func (f *fakeRepo) GetProduct(_ context.Context, _, productID string) (ProductDetail, error) {
	f.lastProductID = productID
	if f.productErr != nil {
		return ProductDetail{}, f.productErr
	}
	return f.createdDetail, nil
}

type fakeUploader struct {
	url    string
	err    error
	called bool
}

func (f *fakeUploader) Upload(_ context.Context, _ string, _ []byte) (string, error) {
	f.called = true
	if f.err != nil {
		return "", f.err
	}
	if f.url == "" {
		return "https://i.ibb.co/hosted/product.png", nil
	}
	return f.url, nil
}

func (f *fakeRepo) ResolveStore(_ context.Context, _, _ string) (StoreScope, error) {
	f.calls++
	if f.scopeErr != nil {
		return StoreScope{}, f.scopeErr
	}
	return f.scope, nil
}

func (f *fakeRepo) Dashboard(_ context.Context, _ string) (Summary, error) {
	if f.summaryErr != nil {
		return Summary{}, f.summaryErr
	}
	return f.summary, nil
}

func (f *fakeRepo) ListProducts(_ context.Context, _ string, q ListQuery) ([]ProductRecord, int64, error) {
	f.lastProductQuery = q
	return f.products, f.productsTotal, nil
}

func (f *fakeRepo) ListOrders(_ context.Context, _ string, q ListQuery) ([]OrderRecord, int64, error) {
	f.lastOrderQuery = q
	return f.orders, f.ordersTotal, nil
}

func (f *fakeRepo) ListInventory(_ context.Context, _ string, q ListQuery) ([]InventoryRecord, int64, error) {
	f.lastInventoryQuery = q
	return f.inventory, f.inventoryTotal, nil
}

func (f *fakeRepo) AdjustStock(_ context.Context, _, variantID, _ string, quantity int, _ string) (InventoryRecord, error) {
	f.lastAdjustVariant, f.lastAdjustQuantity = variantID, quantity
	if f.adjustErr != nil {
		return InventoryRecord{}, f.adjustErr
	}
	return f.adjusted, nil
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		scope: StoreScope{StoreID: "s-1", VendorID: "v-1", Name: "Vendor Demo Co", Status: "active", Role: "owner"},
		products: []ProductRecord{{
			ID: "p-1", Name: "Blender", Slug: "blender", Status: "active",
			ApprovalStatus: "approved", CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		}},
		productsTotal: 1,
		createdDetail: ProductDetail{
			ID: "p-9", Name: "New Product", Slug: "new-product", Status: "draft",
			ApprovalStatus: "pending", Price: 19.99, Stock: 5,
			CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		},
	}
}
