package order

import (
	"context"

	"github.com/md-siam-kazi/e-go-backend/pkg/payment"
)

type fakeRepo struct {
	resolved   []VariantResolution
	resolveErr error

	created        Order
	createErr      error
	lastParams     OrderParams
	createCalls    int
	deletedOrderID string
	deleteOrderErr error

	orders      []Order
	ordersTotal int64
	order       Order
	orderErr    error

	payment         Payment
	paymentErr      error
	paymentCreated  bool
	paymentByRef    Payment
	paymentByRefErr error

	markPaidResult  Payment
	markPaidChanged bool
	markPaidErr     error

	txnInserted  bool
	confirmCalls int
	confirmErr   error
}

func (f *fakeRepo) ResolveVariants(context.Context, []OrderItemInput) ([]VariantResolution, error) {
	return f.resolved, f.resolveErr
}

func (f *fakeRepo) CreateOrder(_ context.Context, p OrderParams) (Order, error) {
	f.createCalls++
	f.lastParams = p
	if f.createErr != nil {
		return Order{}, f.createErr
	}
	created := f.created
	if created.ID == "" {
		created.ID = "order-1"
	}
	if created.OrderNumber == "" {
		created.OrderNumber = p.OrderNumber
	}
	created.Subtotal = p.Subtotal
	created.ShippingTotal = p.ShippingTotal
	created.GrandTotal = p.GrandTotal
	created.Status = StatusPending
	return created, nil
}

func (f *fakeRepo) DeleteOrder(_ context.Context, orderID string) error {
	f.deletedOrderID = orderID
	return f.deleteOrderErr
}

func (f *fakeRepo) OrdersForUser(context.Context, string, int, int) ([]Order, int64, error) {
	return f.orders, f.ordersTotal, nil
}

func (f *fakeRepo) OrderByID(context.Context, string, string) (Order, error) {
	if f.orderErr != nil {
		return Order{}, f.orderErr
	}
	return f.order, nil
}

func (f *fakeRepo) CreatePayment(context.Context, string, string, float64, string) (Payment, error) {
	f.paymentCreated = true
	return f.payment, f.paymentErr
}

func (f *fakeRepo) PaymentByReference(context.Context, string) (Payment, error) {
	return f.paymentByRef, f.paymentByRefErr
}

func (f *fakeRepo) MarkPaid(context.Context, string, string) (Payment, bool, error) {
	return f.markPaidResult, f.markPaidChanged, f.markPaidErr
}

func (f *fakeRepo) InsertPaymentTransaction(context.Context, string, string, float64, string) error {
	f.txnInserted = true
	return nil
}

func (f *fakeRepo) ConfirmPaidOrder(context.Context, string) error {
	f.confirmCalls++
	return f.confirmErr
}

type fakeProvider struct {
	name      string
	paid      bool
	redirect  string
	initErr   error
	verifyErr error
}

func (f *fakeProvider) Name() string { return f.name }

func (f *fakeProvider) Initiate(context.Context, payment.InitRequest) (payment.InitResult, error) {
	if f.initErr != nil {
		return payment.InitResult{}, f.initErr
	}
	return payment.InitResult{Provider: f.name, RedirectURL: f.redirect}, nil
}

func (f *fakeProvider) Verify(context.Context, payment.VerifyRequest) (payment.VerifyResult, error) {
	if f.verifyErr != nil {
		return payment.VerifyResult{}, f.verifyErr
	}
	return payment.VerifyResult{Reference: "SK-1", ProviderTransactionID: "pi_1", Paid: f.paid}, nil
}

func registryWith(p payment.Provider) *payment.Registry {
	r := payment.NewRegistry()
	r.Register(p)
	return r
}

func testConfig(providerName string) Config {
	return Config{
		Currency:              "usd",
		StandardShipping:      5.99,
		ExpressShipping:       12.99,
		FreeShippingThreshold: 75,
		ProviderName:          providerName,
		SuccessURL:            "http://shop.test/order/success",
		CancelURL:             "http://shop.test/checkout",
	}
}

func sampleResolution() VariantResolution {
	return VariantResolution{
		ProductID: "p-1", VariantID: "v-1", VendorID: "vd-1",
		ProductName: "Blender", ProductSKU: "SKU-1", UnitPrice: 39.99, Available: 10,
	}
}

func validRequest(method string) CreateOrderRequest {
	return CreateOrderRequest{
		Email: "buyer@example.com",
		Items: []ItemRequest{{ProductID: "p-1", VariantID: "v-1", Quantity: 1}},
		ShippingAddress: AddressRequest{
			FullName: "Buyer One", Line1: "1 Demo Rd", City: "Dhaka", Country: "Bangladesh",
		},
		DeliveryMethod: DeliveryStandard,
		PaymentMethod:  method,
	}
}
