package vendor

import "time"

// OrderAddress is the buyer's shipping address snapshot for an order.
type OrderAddress struct {
	FullName   string
	Phone      string
	Line1      string
	Line2      string
	City       string
	State      string
	PostalCode string
	Country    string
}

// OrderLineRecord is one of the vendor's lines in an order.
type OrderLineRecord struct {
	OrderItemID string
	ProductID   string
	ProductName string
	ProductSKU  string
	Quantity    int
	UnitPrice   float64
	LineTotal   float64
}

// ShipmentRecord is the vendor's shipment for an order.
type ShipmentRecord struct {
	ID             string
	Carrier        string
	TrackingNumber string
	Status         string
	ShippedAt      *time.Time
	DeliveredAt    *time.Time
	CreatedAt      time.Time
}

// TrackingEventRecord is one entry in a shipment timeline.
type TrackingEventRecord struct {
	ID          string
	Status      string
	Location    string
	Description string
	EventTime   time.Time
	CreatedAt   time.Time
}

// OrderDetailRecord is the full vendor-scoped order view.
type OrderDetailRecord struct {
	OrderID         string
	OrderNumber     string
	Status          string
	Currency        string
	GrandTotal      float64
	CreatedAt       time.Time
	Address         OrderAddress
	Items           []OrderLineRecord
	VendorRevenue   float64
	VendorItemCount int64
	Shipment        *ShipmentRecord
	Events          []TrackingEventRecord
}

// ShipmentParams is a validated shipment/tracking update.
type ShipmentParams struct {
	Carrier        string
	TrackingNumber string
	Status         string
	Location       string
	Description    string
	EventTime      *time.Time
}

// RevenuePoint is one month of the vendor revenue series.
type RevenuePoint struct {
	Month   string
	Revenue float64
	Orders  int64
}

// TopProduct is a best-selling product by revenue.
type TopProduct struct {
	ProductID string
	Name      string
	UnitsSold int64
	Revenue   float64
}

// StatusCount is an order-count/revenue bucket per order status.
type StatusCount struct {
	Status  string
	Count   int64
	Revenue float64
}

// RevenueSummary is the vendor revenue aggregate.
type RevenueSummary struct {
	Revenue           float64
	OrderCount        int64
	UnitsSold         int64
	AverageOrderValue float64
	Series            []RevenuePoint
	TopProducts       []TopProduct
	Statuses          []StatusCount
}
