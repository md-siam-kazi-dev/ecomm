package vendor

import "time"

// StoreScope is the resolved store + backing vendor for a vendor request.
// Vendor operations are store-scoped via store_memberships.
type StoreScope struct {
	StoreID  string
	VendorID string
	Name     string
	Status   string
	Role     string
}

// Summary is the vendor dashboard aggregate, scoped to one vendor.
type Summary struct {
	Revenue           float64
	OrderCount        int64
	ProductCount      int64
	LowStockCount     int64
	PendingFulfilment int64
}

// ProductRecord is a product owned by the vendor (list view).
type ProductRecord struct {
	ID             string
	Name           string
	Slug           string
	Status         string
	ApprovalStatus string
	ReviewNote     string
	Price          float64
	Stock          int64
	ImageURL       string
	SubmittedAt    *time.Time
	ReviewedAt     *time.Time
	CreatedAt      time.Time
}

// ProductImage is one stored product image.
type ProductImage struct {
	URL       string
	AltText   string
	SortOrder int
	IsPrimary bool
}

// ProductDetail is the full product view (vendor edit / admin review).
type ProductDetail struct {
	ID               string
	Name             string
	Slug             string
	Description      string
	ShortDescription string
	Status           string
	ApprovalStatus   string
	Category         string
	Brand            string
	SKU              string
	Price            float64
	SalePrice        *float64
	Stock            int64
	LowStock         int
	Images           []ProductImage
	Attributes       map[string]string
	ReviewNote       string
	VendorName       string
	SubmittedAt      *time.Time
	ApprovedAt       *time.Time
	ReviewedAt       *time.Time
	CreatedAt        time.Time
}

// ProductParams is the validated create/update payload.
type ProductParams struct {
	Name              string
	Description       string
	ShortDescription  string
	CategorySlug      string
	Brand             string
	SKU               string
	Price             float64
	SalePrice         *float64
	Stock             int
	LowStockThreshold int
	Images            []string
	Attributes        map[string]string
	// Submit marks the product as submitted for review; otherwise it stays a
	// draft (still pending approval, but not in the review queue).
	Submit bool
}

// OrderRecord is an order (summarized) that contains the vendor's items.
type OrderRecord struct {
	OrderID     string
	OrderNumber string
	Status      string
	ItemCount   int64
	Total       float64
	CreatedAt   time.Time
}

// InventoryRecord is one variant's stock for the vendor inventory list.
type InventoryRecord struct {
	VariantID    string
	ProductID    string
	Name         string
	SKU          string
	Category     string
	ImageURL     string
	Quantity     int64
	Reserved     int64
	ReorderLevel int64
}

// InventoryMovementRecord is one entry in the vendor's stock history.
type InventoryMovementRecord struct {
	ID             string
	VariantID      string
	ProductID      string
	Name           string
	SKU            string
	MovementType   string
	Quantity       int64
	QuantityBefore *int64
	QuantityAfter  *int64
	Notes          string
	ReferenceType  string
	OrderNumber    string
	Actor          string
	CreatedAt      time.Time
}

// ExportQuery filters an export. Zero From/To mean "unbounded".
type ExportQuery struct {
	Status string
	From   time.Time
	To     time.Time
}

// ExportOrderRow is one order-item line in the orders export (one row per item).
type ExportOrderRow struct {
	OrderNumber       string
	OrderDate         time.Time
	ProductName       string
	SKU               string
	Quantity          int
	UnitPrice         float64
	Discount          float64
	LineTotal         float64
	PaymentStatus     string
	FulfillmentStatus string
	TrackingNumber    string
	Currency          string
}

// ExportInventoryRow is one variant line in the inventory export.
type ExportInventoryRow struct {
	ProductID      string
	Name           string
	SKU            string
	Price          float64
	CostPrice      *float64
	Quantity       int64
	Reserved       int64
	ReorderLevel   int64
	ProductStatus  string
	ApprovalStatus string
	UpdatedAt      *time.Time
}
