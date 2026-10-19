package order

import "time"

// ItemRequest is a cart line from the client (ids + quantity only).
type ItemRequest struct {
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

// AddressRequest is the submitted shipping address.
type AddressRequest struct {
	FullName   string `json:"full_name"`
	Phone      string `json:"phone"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// CreateOrderRequest is the checkout payload.
type CreateOrderRequest struct {
	Email           string         `json:"email"`
	Items           []ItemRequest  `json:"items"`
	ShippingAddress AddressRequest `json:"shipping_address"`
	DeliveryMethod  string         `json:"delivery_method"`
	PaymentMethod   string         `json:"payment_method"`
}

// OrderItemReviewResponse is the caller's existing review for an order line.
type OrderItemReviewResponse struct {
	ID        string   `json:"id"`
	Rating    int      `json:"rating"`
	Title     string   `json:"title"`
	Comment   string   `json:"comment"`
	Images    []string `json:"images,omitempty"`
	CreatedAt string   `json:"created_at"`
}

// OrderItemResponse is one line in an order.
type OrderItemResponse struct {
	OrderItemID string                   `json:"order_item_id"`
	ProductID   string                   `json:"product_id"`
	VariantID   string                   `json:"variant_id,omitempty"`
	ProductName string                   `json:"product_name"`
	ProductSKU  string                   `json:"product_sku,omitempty"`
	Quantity    int                      `json:"quantity"`
	UnitPrice   float64                  `json:"unit_price"`
	LineTotal   float64                  `json:"line_total"`
	ImageURL    string                   `json:"image_url,omitempty"`
	Reviewed    bool                     `json:"reviewed"`
	Review      *OrderItemReviewResponse `json:"review,omitempty"`
}

// OrderResponse is a persisted order.
type OrderResponse struct {
	ID            string              `json:"id"`
	OrderNumber   string              `json:"order_number"`
	Status        string              `json:"status"`
	Currency      string              `json:"currency"`
	Subtotal      float64             `json:"subtotal"`
	ShippingTotal float64             `json:"shipping_total"`
	TaxTotal      float64             `json:"tax_total"`
	GrandTotal    float64             `json:"grand_total"`
	CreatedAt     string              `json:"created_at"`
	Items         []OrderItemResponse `json:"items"`
	Shipping      AddressRequest      `json:"shipping_address"`
}

// CreateOrderResponse is returned by POST /orders.
type CreateOrderResponse struct {
	Order       OrderResponse `json:"order"`
	CheckoutURL string        `json:"checkout_url,omitempty"`
}

func toOrderResponse(o Order) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items))
	for _, item := range o.Items {
		resp := OrderItemResponse{
			OrderItemID: item.OrderItemID,
			ProductID:   item.ProductID,
			VariantID:   item.VariantID,
			ProductName: item.ProductName,
			ProductSKU:  item.ProductSKU,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			LineTotal:   item.LineTotal,
			ImageURL:    item.ImageURL,
			Reviewed:    item.Reviewed,
		}
		if item.Review != nil {
			resp.Review = &OrderItemReviewResponse{
				ID:        item.Review.ID,
				Rating:    item.Review.Rating,
				Title:     item.Review.Title,
				Comment:   item.Review.Comment,
				Images:    item.Review.Images,
				CreatedAt: item.Review.CreatedAt.UTC().Format(time.RFC3339),
			}
		}
		items = append(items, resp)
	}

	return OrderResponse{
		ID:            o.ID,
		OrderNumber:   o.OrderNumber,
		Status:        o.Status,
		Currency:      o.Currency,
		Subtotal:      o.Subtotal,
		ShippingTotal: o.ShippingTotal,
		TaxTotal:      o.TaxTotal,
		GrandTotal:    o.GrandTotal,
		CreatedAt:     o.CreatedAt.UTC().Format(time.RFC3339),
		Items:         items,
		Shipping: AddressRequest{
			FullName:   o.Shipping.FullName,
			Phone:      o.Shipping.Phone,
			Line1:      o.Shipping.Line1,
			Line2:      o.Shipping.Line2,
			City:       o.Shipping.City,
			State:      o.Shipping.State,
			PostalCode: o.Shipping.PostalCode,
			Country:    o.Shipping.Country,
		},
	}
}
