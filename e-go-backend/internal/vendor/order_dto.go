package vendor

import "time"

// OrderAddressResponse is the buyer's shipping address snapshot.
type OrderAddressResponse struct {
	FullName   string `json:"full_name"`
	Phone      string `json:"phone,omitempty"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country"`
}

// OrderLineResponse is one of the vendor's lines in an order.
type OrderLineResponse struct {
	OrderItemID string  `json:"order_item_id"`
	ProductID   string  `json:"product_id,omitempty"`
	Name        string  `json:"name"`
	SKU         string  `json:"sku,omitempty"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	LineTotal   float64 `json:"line_total"`
}

// ShipmentResponse is the vendor's shipment for an order.
type ShipmentResponse struct {
	ID             string  `json:"id"`
	Carrier        string  `json:"carrier,omitempty"`
	TrackingNumber string  `json:"tracking_number,omitempty"`
	TrackingURL    string  `json:"tracking_url,omitempty"`
	Status         string  `json:"status"`
	ShippedAt      *string `json:"shipped_at,omitempty"`
	DeliveredAt    *string `json:"delivered_at,omitempty"`
}

// TrackingEventResponse is one entry in the shipment timeline.
type TrackingEventResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
	EventTime   string `json:"event_time"`
}

// OrderDetailResponse is the full vendor-scoped order view.
type OrderDetailResponse struct {
	ID              string                  `json:"id"`
	OrderNumber     string                  `json:"order_number"`
	Status          string                  `json:"status"`
	Currency        string                  `json:"currency"`
	GrandTotal      float64                 `json:"grand_total"`
	CreatedAt       string                  `json:"created_at"`
	ShippingAddress OrderAddressResponse    `json:"shipping_address"`
	Items           []OrderLineResponse     `json:"items"`
	VendorRevenue   float64                 `json:"vendor_revenue"`
	VendorItemCount int64                   `json:"vendor_item_count"`
	Shipment        *ShipmentResponse       `json:"shipment,omitempty"`
	Events          []TrackingEventResponse `json:"events"`
}

// ShipmentEventRequest is a tracking update posted by the vendor.
type ShipmentEventRequest struct {
	Carrier        string `json:"carrier"`
	TrackingNumber string `json:"tracking_number"`
	Status         string `json:"status"`
	Location       string `json:"location"`
	Description    string `json:"description"`
	EventTime      string `json:"event_time"`
}

// OrderStatusRequest is an order-level fulfilment transition.
type OrderStatusRequest struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

// RevenuePointResponse is one month of the revenue series.
type RevenuePointResponse struct {
	Month   string  `json:"month"`
	Revenue float64 `json:"revenue"`
	Orders  int64   `json:"orders"`
}

// TopProductResponse is a best-selling product by revenue.
type TopProductResponse struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	UnitsSold int64   `json:"units_sold"`
	Revenue   float64 `json:"revenue"`
}

// StatusCountResponse is an order count/revenue bucket per status.
type StatusCountResponse struct {
	Status  string  `json:"status"`
	Count   int64   `json:"count"`
	Revenue float64 `json:"revenue"`
}

// RevenueResponse is the vendor revenue aggregate.
type RevenueResponse struct {
	Revenue           float64                `json:"revenue"`
	OrderCount        int64                  `json:"order_count"`
	UnitsSold         int64                  `json:"units_sold"`
	AverageOrderValue float64                `json:"average_order_value"`
	Series            []RevenuePointResponse `json:"series"`
	TopProducts       []TopProductResponse   `json:"top_products"`
	Statuses          []StatusCountResponse  `json:"statuses"`
}

func toOrderDetailResponse(d OrderDetailRecord) OrderDetailResponse {
	items := make([]OrderLineResponse, 0, len(d.Items))
	for _, line := range d.Items {
		items = append(items, OrderLineResponse{
			OrderItemID: line.OrderItemID,
			ProductID:   line.ProductID,
			Name:        line.ProductName,
			SKU:         line.ProductSKU,
			Quantity:    line.Quantity,
			UnitPrice:   line.UnitPrice,
			LineTotal:   line.LineTotal,
		})
	}

	events := make([]TrackingEventResponse, 0, len(d.Events))
	for _, ev := range d.Events {
		events = append(events, TrackingEventResponse{
			ID:          ev.ID,
			Status:      ev.Status,
			Location:    ev.Location,
			Description: ev.Description,
			EventTime:   ev.EventTime.UTC().Format(time.RFC3339),
		})
	}

	resp := OrderDetailResponse{
		ID:          d.OrderID,
		OrderNumber: d.OrderNumber,
		Status:      d.Status,
		Currency:    d.Currency,
		GrandTotal:  d.GrandTotal,
		CreatedAt:   d.CreatedAt.UTC().Format(time.RFC3339),
		ShippingAddress: OrderAddressResponse{
			FullName:   d.Address.FullName,
			Phone:      d.Address.Phone,
			Line1:      d.Address.Line1,
			Line2:      d.Address.Line2,
			City:       d.Address.City,
			State:      d.Address.State,
			PostalCode: d.Address.PostalCode,
			Country:    d.Address.Country,
		},
		Items:           items,
		VendorRevenue:   d.VendorRevenue,
		VendorItemCount: d.VendorItemCount,
		Events:          events,
	}
	if d.Shipment != nil {
		resp.Shipment = &ShipmentResponse{
			ID:             d.Shipment.ID,
			Carrier:        d.Shipment.Carrier,
			TrackingNumber: d.Shipment.TrackingNumber,
			Status:         d.Shipment.Status,
			ShippedAt:      fmtTime(d.Shipment.ShippedAt),
			DeliveredAt:    fmtTime(d.Shipment.DeliveredAt),
		}
	}
	return resp
}

func toRevenueResponse(s RevenueSummary) RevenueResponse {
	series := make([]RevenuePointResponse, 0, len(s.Series))
	for _, p := range s.Series {
		series = append(series, RevenuePointResponse{Month: p.Month, Revenue: p.Revenue, Orders: p.Orders})
	}
	top := make([]TopProductResponse, 0, len(s.TopProducts))
	for _, p := range s.TopProducts {
		top = append(top, TopProductResponse{ProductID: p.ProductID, Name: p.Name, UnitsSold: p.UnitsSold, Revenue: p.Revenue})
	}
	statuses := make([]StatusCountResponse, 0, len(s.Statuses))
	for _, c := range s.Statuses {
		statuses = append(statuses, StatusCountResponse{Status: c.Status, Count: c.Count, Revenue: c.Revenue})
	}
	return RevenueResponse{
		Revenue:           s.Revenue,
		OrderCount:        s.OrderCount,
		UnitsSold:         s.UnitsSold,
		AverageOrderValue: s.AverageOrderValue,
		Series:            series,
		TopProducts:       top,
		Statuses:          statuses,
	}
}
