package order

import (
	"net/http"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// Order statuses (subset of the order_status enum we drive).
const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
)

// Payment methods.
const (
	MethodCard = "card"
	MethodCOD  = "cod"
)

// Delivery methods.
const (
	DeliveryStandard = "standard"
	DeliveryExpress  = "express"
)

var (
	errInvalidBody       = apperr.New(http.StatusBadRequest, "BAD_REQUEST", "Request body must be valid JSON")
	errNotFound          = apperr.New(http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
	errEmptyItems        = apperr.New(http.StatusUnprocessableEntity, "EMPTY_ORDER", "Your cart is empty")
	errNoAddress         = apperr.New(http.StatusUnprocessableEntity, "VALIDATION_ERROR", "A shipping address is required")
	errBadDelivery       = apperr.New(http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Choose a valid delivery method")
	errBadPayment        = apperr.New(http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Choose a valid payment method")
	errUnavailable       = apperr.New(http.StatusConflict, "PRODUCT_UNAVAILABLE", "A product in your cart is no longer available")
	errInsufficientStock = apperr.New(http.StatusConflict, "INSUFFICIENT_STOCK", "A product in your cart does not have enough stock")
	// errPaymentNotConfigured is returned when online payment is requested but no provider is set.
	errPaymentNotConfigured = apperr.New(http.StatusServiceUnavailable, "PAYMENT_NOT_CONFIGURED", "Online payment is not available yet — no payment provider is configured")
)

// OrderItemInput is a requested line: ids + quantity only (never a price).
type OrderItemInput struct {
	ProductID string
	VariantID string
	Quantity  int
}

// VariantResolution is a validated line resolved from the catalog.
type VariantResolution struct {
	ProductID   string
	VariantID   string
	VendorID    string
	ProductName string
	ProductSKU  string
	UnitPrice   float64
	Available   int64
}

// AddressParams is a shipping address snapshot.
type AddressParams struct {
	FullName   string
	Phone      string
	Line1      string
	Line2      string
	City       string
	State      string
	PostalCode string
	Country    string
}

// OrderParams is the validated, server-priced order ready to persist.
type OrderParams struct {
	UserID        string
	OrderNumber   string
	Currency      string
	Items         []VariantResolution
	Quantities    []int
	Subtotal      float64
	ShippingTotal float64
	TaxTotal      float64
	GrandTotal    float64
	Shipping      AddressParams
}

// Order is a persisted order with its items.
type Order struct {
	ID            string
	OrderNumber   string
	Status        string
	Currency      string
	Subtotal      float64
	ShippingTotal float64
	TaxTotal      float64
	GrandTotal    float64
	CreatedAt     time.Time
	Items         []OrderItem
	Shipping      AddressParams
}

// OrderItem is one persisted line snapshot.
type OrderItem struct {
	OrderItemID string
	ProductID   string
	VariantID   string
	ProductName string
	ProductSKU  string
	Quantity    int
	UnitPrice   float64
	LineTotal   float64
	ImageURL    string
	Reviewed    bool
	Review      *OrderItemReview
}

// OrderItemReview is the caller's existing review for an order line, used to
// prefill the edit form on the order details page.
type OrderItemReview struct {
	ID        string
	Rating    int
	Title     string
	Comment   string
	CreatedAt time.Time
}

// Payment is an order payment row.
type Payment struct {
	ID                string
	OrderID           string
	Reference         string // the order number
	Provider          string
	ProviderPaymentID string
	Status            string
	Amount            float64
	Currency          string
	VerifiedAt        *time.Time
}
