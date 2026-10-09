package admin

import "time"

// ReviewProductResponse is a product row in the admin review queue.
type ReviewProductResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	VendorName     string  `json:"vendor_name"`
	Status         string  `json:"status"`
	ApprovalStatus string  `json:"approval_status"`
	Price          float64 `json:"price"`
	Stock          int64   `json:"stock"`
	ImageURL       string  `json:"image_url,omitempty"`
	ReviewNote     string  `json:"review_note,omitempty"`
	SubmittedAt    *string `json:"submitted_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

// ReviewImageResponse is one product image.
type ReviewImageResponse struct {
	URL       string `json:"url"`
	Alt       string `json:"alt,omitempty"`
	IsPrimary bool   `json:"is_primary"`
	SortOrder int    `json:"sort_order"`
}

// ReviewProductDetailResponse is the full product view for moderation.
type ReviewProductDetailResponse struct {
	ID               string                `json:"id"`
	Name             string                `json:"name"`
	Slug             string                `json:"slug"`
	Description      string                `json:"description"`
	ShortDescription string                `json:"short_description"`
	Status           string                `json:"status"`
	ApprovalStatus   string                `json:"approval_status"`
	Category         string                `json:"category"`
	Brand            string                `json:"brand,omitempty"`
	SKU              string                `json:"sku,omitempty"`
	Price            float64               `json:"price"`
	SalePrice        *float64              `json:"sale_price,omitempty"`
	Stock            int64                 `json:"stock"`
	LowStock         int                   `json:"low_stock_threshold"`
	Images           []ReviewImageResponse `json:"images"`
	Attributes       map[string]string     `json:"attributes"`
	ReviewNote       string                `json:"review_note,omitempty"`
	VendorName       string                `json:"vendor_name,omitempty"`
	SubmittedAt      *string               `json:"submitted_at,omitempty"`
	ApprovedAt       *string               `json:"approved_at,omitempty"`
	CreatedAt        string                `json:"created_at"`
}

// ReviewDecisionRequest carries the rejection reason / change request note.
type ReviewDecisionRequest struct {
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

func fmtTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

func toReviewProductResponse(r ReviewProduct) ReviewProductResponse {
	return ReviewProductResponse{
		ID:             r.ID,
		Name:           r.Name,
		VendorName:     r.VendorName,
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

func toReviewProductDetailResponse(d ReviewProductDetail) ReviewProductDetailResponse {
	images := make([]ReviewImageResponse, 0, len(d.Images))
	for _, img := range d.Images {
		images = append(images, ReviewImageResponse{URL: img.URL, Alt: img.Alt, IsPrimary: img.IsPrimary, SortOrder: img.SortOrder})
	}
	if d.Attributes == nil {
		d.Attributes = map[string]string{}
	}
	return ReviewProductDetailResponse{
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
