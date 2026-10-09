package vendor

import "time"

// ListQuery is the normalized filter for a vendor list endpoint.
type ListQuery struct {
	Page   int
	Limit  int
	Search string
	Status string
}

// ProductRequest is the create/update payload for a product. Any status,
// approval or vendor fields sent by the client are ignored.
type ProductRequest struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	ShortDescription  string            `json:"short_description"`
	Category          string            `json:"category"` // category slug
	Brand             string            `json:"brand"`
	SKU               string            `json:"sku"`
	Price             float64           `json:"price"`
	SalePrice         *float64          `json:"sale_price"`
	Stock             int               `json:"stock"`
	LowStockThreshold int               `json:"low_stock_threshold"`
	Images            []string          `json:"images"`
	Attributes        map[string]string `json:"attributes"`
	Submit            bool              `json:"submit"`
}

// ProductImageResponse is one product image.
type ProductImageResponse struct {
	URL       string `json:"url"`
	Alt       string `json:"alt,omitempty"`
	IsPrimary bool   `json:"is_primary"`
	SortOrder int    `json:"sort_order"`
}

// ProductItem is a row in the vendor's product list.
type ProductItem struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	Status         string  `json:"status"`
	ApprovalStatus string  `json:"approval_status"`
	Price          float64 `json:"price"`
	Stock          int64   `json:"stock"`
	ImageURL       string  `json:"image_url,omitempty"`
	ReviewNote     string  `json:"review_note,omitempty"`
	SubmittedAt    *string `json:"submitted_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

// ProductDetailResponse is the full product view.
type ProductDetailResponse struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Slug             string                 `json:"slug"`
	Description      string                 `json:"description"`
	ShortDescription string                 `json:"short_description"`
	Status           string                 `json:"status"`
	ApprovalStatus   string                 `json:"approval_status"`
	Category         string                 `json:"category"`
	Brand            string                 `json:"brand,omitempty"`
	SKU              string                 `json:"sku,omitempty"`
	Price            float64                `json:"price"`
	SalePrice        *float64               `json:"sale_price,omitempty"`
	Stock            int64                  `json:"stock"`
	LowStock         int                    `json:"low_stock_threshold"`
	Images           []ProductImageResponse `json:"images"`
	Attributes       map[string]string      `json:"attributes"`
	ReviewNote       string                 `json:"review_note,omitempty"`
	VendorName       string                 `json:"vendor_name,omitempty"`
	SubmittedAt      *string                `json:"submitted_at,omitempty"`
	ApprovedAt       *string                `json:"approved_at,omitempty"`
	CreatedAt        string                 `json:"created_at"`
}

// ImageUploadResponse is returned after hosting a product image.
type ImageUploadResponse struct {
	URL string `json:"url"`
}

// ProfileResponse is the vendor's own store profile.
type ProfileResponse struct {
	ID           string `json:"id"`
	StoreID      string `json:"store_id"`
	BusinessName string `json:"business_name"`
	Status       string `json:"status"`
	Role         string `json:"role"`
}

// SummaryResponse is the vendor dashboard aggregate.
type SummaryResponse struct {
	Revenue           float64 `json:"revenue"`
	OrderCount        int64   `json:"order_count"`
	ProductCount      int64   `json:"product_count"`
	LowStockCount     int64   `json:"low_stock_count"`
	PendingFulfilment int64   `json:"pending_fulfilment"`
}

// OrderItem is a row in the vendor's order list.
type OrderItem struct {
	OrderID     string  `json:"order_id"`
	OrderNumber string  `json:"order_number"`
	Status      string  `json:"status"`
	ItemCount   int64   `json:"item_count"`
	Total       float64 `json:"total"`
	CreatedAt   string  `json:"created_at"`
}

// InventoryItem is a row in the vendor's inventory list (per variant).
type InventoryItem struct {
	VariantID    string `json:"variant_id"`
	ProductID    string `json:"product_id"`
	Name         string `json:"name"`
	SKU          string `json:"sku"`
	Category     string `json:"category"`
	ImageURL     string `json:"image_url,omitempty"`
	Quantity     int64  `json:"quantity"`
	Reserved     int64  `json:"reserved"`
	Available    int64  `json:"available"`
	ReorderLevel int64  `json:"reorder_level"`
}

// InventoryAdjustRequest sets a variant's on-hand stock.
type InventoryAdjustRequest struct {
	Quantity int    `json:"quantity"`
	Note     string `json:"note"`
}

func toInventoryItem(r InventoryRecord) InventoryItem {
	return InventoryItem{
		VariantID:    r.VariantID,
		ProductID:    r.ProductID,
		Name:         r.Name,
		SKU:          r.SKU,
		Category:     r.Category,
		ImageURL:     r.ImageURL,
		Quantity:     r.Quantity,
		Reserved:     r.Reserved,
		Available:    r.Quantity - r.Reserved,
		ReorderLevel: r.ReorderLevel,
	}
}

func fmtTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

func toProductItem(r ProductRecord) ProductItem {
	return ProductItem{
		ID:             r.ID,
		Name:           r.Name,
		Slug:           r.Slug,
		Status:         r.Status,
		ApprovalStatus: r.ApprovalStatus,
		Price:          r.Price,
		Stock:          r.Stock,
		ImageURL:       r.ImageURL,
		ReviewNote:     r.ReviewNote,
		SubmittedAt:    fmtTime(r.SubmittedAt),
		CreatedAt:      r.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toProductDetailResponse(d ProductDetail) ProductDetailResponse {
	images := make([]ProductImageResponse, 0, len(d.Images))
	for _, img := range d.Images {
		images = append(images, ProductImageResponse{
			URL: img.URL, Alt: img.AltText, IsPrimary: img.IsPrimary, SortOrder: img.SortOrder,
		})
	}
	if d.Attributes == nil {
		d.Attributes = map[string]string{}
	}
	return ProductDetailResponse{
		ID:               d.ID,
		Name:             d.Name,
		Slug:             d.Slug,
		Description:      d.Description,
		ShortDescription: d.ShortDescription,
		Status:           d.Status,
		ApprovalStatus:   d.ApprovalStatus,
		Category:         d.Category,
		Brand:            d.Brand,
		SKU:              d.SKU,
		Price:            d.Price,
		SalePrice:        d.SalePrice,
		Stock:            d.Stock,
		LowStock:         d.LowStock,
		Images:           images,
		Attributes:       d.Attributes,
		ReviewNote:       d.ReviewNote,
		VendorName:       d.VendorName,
		SubmittedAt:      fmtTime(d.SubmittedAt),
		ApprovedAt:       fmtTime(d.ApprovedAt),
		CreatedAt:        d.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toSummaryResponse(s Summary) SummaryResponse {
	return SummaryResponse{
		Revenue:           s.Revenue,
		OrderCount:        s.OrderCount,
		ProductCount:      s.ProductCount,
		LowStockCount:     s.LowStockCount,
		PendingFulfilment: s.PendingFulfilment,
	}
}

func toOrderItem(r OrderRecord) OrderItem {
	return OrderItem{
		OrderID:     r.OrderID,
		OrderNumber: r.OrderNumber,
		Status:      r.Status,
		ItemCount:   r.ItemCount,
		Total:       r.Total,
		CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
	}
}
