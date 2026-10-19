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

	orderDetail       OrderDetailRecord
	orderDetailErr    error
	lastOrderDetailID string
	lastShipment      ShipmentParams
	shipmentErr       error
	lastOrderStatus   string
	orderStatusErr    error
	revenue           RevenueSummary
	revenueErr        error

	movements          []InventoryMovementRecord
	movementsTotal     int64
	lastMovementsQuery ListQuery
	thresholdRecord    InventoryRecord
	thresholdErr       error
	lastThresholdVar   string
	lastThresholdLevel int

	exportOrders    []ExportOrderRow
	exportInventory []ExportInventoryRow

	reviews          []VendorReviewRecord
	reviewsTotal     int64
	lastReviewQuery  ListQuery
	reviewSummary    ReviewSummaryRecord
	reviewSummaryErr error
	replyRecord      VendorReviewRecord
	replyErr         error
	lastReply        string

	questions         []VendorQuestionRecord
	questionsTotal    int64
	lastQuestionQuery ListQuery
	answered          VendorQuestionRecord
	answerErr         error
	lastAnswer        string
}

func (f *fakeRepo) OrderDetail(_ context.Context, _, orderID string) (OrderDetailRecord, error) {
	f.lastOrderDetailID = orderID
	if f.orderDetailErr != nil {
		return OrderDetailRecord{}, f.orderDetailErr
	}
	return f.orderDetail, nil
}

func (f *fakeRepo) UpsertShipmentEvent(_ context.Context, _, orderID, _ string, p ShipmentParams) error {
	f.lastShipment, f.lastOrderDetailID = p, orderID
	return f.shipmentErr
}

func (f *fakeRepo) SetOrderStatus(_ context.Context, _, orderID, _, status, _ string) error {
	f.lastOrderStatus, f.lastOrderDetailID = status, orderID
	return f.orderStatusErr
}

func (f *fakeRepo) ListInventoryMovements(_ context.Context, _ string, q ListQuery) ([]InventoryMovementRecord, int64, error) {
	f.lastMovementsQuery = q
	return f.movements, f.movementsTotal, nil
}

func (f *fakeRepo) SetReorderLevel(_ context.Context, _, variantID, _ string, level int) (InventoryRecord, error) {
	f.lastThresholdVar, f.lastThresholdLevel = variantID, level
	if f.thresholdErr != nil {
		return InventoryRecord{}, f.thresholdErr
	}
	return f.thresholdRecord, nil
}

func (f *fakeRepo) ExportOrders(_ context.Context, _ string, _ ExportQuery) ([]ExportOrderRow, error) {
	return f.exportOrders, nil
}

func (f *fakeRepo) ExportInventory(_ context.Context, _ string, _ ListQuery) ([]ExportInventoryRow, error) {
	return f.exportInventory, nil
}

func (f *fakeRepo) Revenue(_ context.Context, _ string) (RevenueSummary, error) {
	if f.revenueErr != nil {
		return RevenueSummary{}, f.revenueErr
	}
	return f.revenue, nil
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

func (f *fakeRepo) ListReviews(_ context.Context, _ string, q ListQuery) ([]VendorReviewRecord, int64, error) {
	f.lastReviewQuery = q
	return f.reviews, f.reviewsTotal, nil
}

func (f *fakeRepo) ReviewSummary(_ context.Context, _ string) (ReviewSummaryRecord, error) {
	if f.reviewSummaryErr != nil {
		return ReviewSummaryRecord{}, f.reviewSummaryErr
	}
	return f.reviewSummary, nil
}

func (f *fakeRepo) ReplyToReview(_ context.Context, _, _ string, reply string) (VendorReviewRecord, error) {
	f.lastReply = reply
	if f.replyErr != nil {
		return VendorReviewRecord{}, f.replyErr
	}
	return f.replyRecord, nil
}

func (f *fakeRepo) ListQuestions(_ context.Context, _ string, q ListQuery) ([]VendorQuestionRecord, int64, error) {
	f.lastQuestionQuery = q
	return f.questions, f.questionsTotal, nil
}

func (f *fakeRepo) AnswerQuestion(_ context.Context, _, _, _ string, answer string) (VendorQuestionRecord, error) {
	f.lastAnswer = answer
	if f.answerErr != nil {
		return VendorQuestionRecord{}, f.answerErr
	}
	return f.answered, nil
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
