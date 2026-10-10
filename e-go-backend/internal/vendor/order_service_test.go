package vendor

import (
	"context"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

func TestAddShipmentEvent_RejectsBadStatus(t *testing.T) {
	_, err := NewService(newFakeRepo(), &fakeUploader{}).AddShipmentEvent(
		context.Background(), "u-1", "", "o-1", ShipmentEventRequest{Status: "teleported"})

	appErr, ok := apperr.As(err)
	if !ok || appErr.Code != errInvalidShipment.Code {
		t.Fatalf("expected invalid shipment error, got %v", err)
	}
}

func TestAddShipmentEvent_NormalizesAndReturnsDetail(t *testing.T) {
	repo := newFakeRepo()
	repo.orderDetail = OrderDetailRecord{OrderID: "o-1", OrderNumber: "SK-1", Status: "shipped"}

	got, err := NewService(repo, &fakeUploader{}).AddShipmentEvent(
		context.Background(), "u-1", "", "o-1",
		ShipmentEventRequest{Status: "SHIPPED", Carrier: "  DHL  ", TrackingNumber: " T1 ", Description: "  out for delivery  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastShipment.Status != "shipped" || repo.lastShipment.Carrier != "DHL" || repo.lastShipment.Description != "out for delivery" {
		t.Fatalf("shipment not normalized: %+v", repo.lastShipment)
	}
	if repo.lastOrderDetailID != "o-1" || got.OrderNumber != "SK-1" {
		t.Fatalf("order not refreshed: id=%q got=%+v", repo.lastOrderDetailID, got)
	}
}

func TestOrderDetail_MapsRecord(t *testing.T) {
	repo := newFakeRepo()
	repo.orderDetail = OrderDetailRecord{
		OrderID: "o-1", OrderNumber: "SK-9", Status: "processing", Currency: "usd", GrandTotal: 45.98,
		Items:         []OrderLineRecord{{OrderItemID: "oi-1", ProductName: "Blender", Quantity: 2, UnitPrice: 39.99, LineTotal: 79.98}},
		VendorRevenue: 79.98, VendorItemCount: 2,
	}

	got, err := NewService(repo, &fakeUploader{}).OrderDetail(context.Background(), "u-1", "", "o-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.OrderNumber != "SK-9" || len(got.Items) != 1 || got.Items[0].Name != "Blender" || got.VendorRevenue != 79.98 {
		t.Fatalf("unexpected detail: %+v", got)
	}
}

func TestRevenue_ReturnsAggregate(t *testing.T) {
	repo := newFakeRepo()
	repo.revenue = RevenueSummary{
		Revenue: 149.97, OrderCount: 3, UnitsSold: 5, AverageOrderValue: 49.99,
		Series:      []RevenuePoint{{Month: "2026-05", Revenue: 149.97, Orders: 3}},
		TopProducts: []TopProduct{{ProductID: "p-1", Name: "Blender", UnitsSold: 5, Revenue: 149.97}},
		Statuses:    []StatusCount{{Status: "delivered", Count: 2, Revenue: 99.98}},
	}

	got, err := NewService(repo, &fakeUploader{}).Revenue(context.Background(), "u-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Revenue != 149.97 || len(got.Series) != 1 || len(got.TopProducts) != 1 || got.TopProducts[0].Name != "Blender" {
		t.Fatalf("unexpected revenue: %+v", got)
	}
}
