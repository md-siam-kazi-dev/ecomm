package cart

// AddItemRequest is the add-to-cart payload. variant_id may be empty, in which
// case the server resolves the product's default variant.
type AddItemRequest struct {
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

// UpdateQuantityRequest sets a line's quantity (0 removes the line).
type UpdateQuantityRequest struct {
	Quantity *int `json:"quantity"`
}

// CartItemResponse is one display-ready cart line.
type CartItemResponse struct {
	ID           string  `json:"id"`
	ProductID    string  `json:"product_id"`
	VariantID    string  `json:"variant_id"`
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	ImageURL     string  `json:"image_url"`
	ImageAlt     string  `json:"image_alt"`
	VariantLabel string  `json:"variant_label,omitempty"`
	UnitPrice    float64 `json:"unit_price"`
	Quantity     int     `json:"quantity"`
	LineTotal    float64 `json:"line_total"`
	MaxQuantity  int     `json:"max_quantity"`
	InStock      bool    `json:"in_stock"`
}

// CartResponse is the full cart returned after every read or mutation, so the
// client can treat the server as the single source of truth.
type CartResponse struct {
	ID        string             `json:"id"`
	Currency  string             `json:"currency"`
	ItemCount int                `json:"item_count"`
	Subtotal  float64            `json:"subtotal"`
	Items     []CartItemResponse `json:"items"`
}

// toResponse projects a cart + its persisted lines into the API shape.
func toResponse(c Cart, records []ItemRecord) CartResponse {
	items := make([]CartItemResponse, 0, len(records))
	itemCount := 0
	subtotal := 0.0
	for _, r := range records {
		lineTotal := round2(r.UnitPrice * float64(r.Quantity))
		itemCount += r.Quantity
		subtotal += lineTotal

		maxQuantity := int(r.Available)
		if maxQuantity < r.Quantity {
			// Never report a max below the quantity the shopper already has.
			maxQuantity = r.Quantity
		}
		items = append(items, CartItemResponse{
			ID:           r.ID,
			ProductID:    r.ProductID,
			VariantID:    r.VariantID,
			Name:         r.Name,
			Slug:         r.Slug,
			ImageURL:     r.ImageURL,
			ImageAlt:     r.ImageAlt,
			VariantLabel: r.VariantLabel,
			UnitPrice:    round2(r.UnitPrice),
			Quantity:     r.Quantity,
			LineTotal:    lineTotal,
			MaxQuantity:  maxQuantity,
			InStock:      r.Available > 0,
		})
	}
	return CartResponse{
		ID:        c.ID,
		Currency:  c.Currency,
		ItemCount: itemCount,
		Subtotal:  round2(subtotal),
		Items:     items,
	}
}
