package vendor

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

func TestSetOrderStatus_RejectsUnknownStatus(t *testing.T) {
	_, err := NewService(newFakeRepo(), &fakeUploader{}).SetOrderStatus(
		context.Background(), "u-1", "", "o-1", OrderStatusRequest{Status: "teleported"})

	appErr, ok := apperr.As(err)
	if !ok || appErr.Code != errInvalidTransition.Code {
		t.Fatalf("expected invalid transition, got %v", err)
	}
}

func TestSetOrderStatus_NormalizesAndReturns(t *testing.T) {
	repo := newFakeRepo()
	repo.orderDetail = OrderDetailRecord{OrderID: "o-1", OrderNumber: "SK-2", Status: "confirmed"}

	got, err := NewService(repo, &fakeUploader{}).SetOrderStatus(
		context.Background(), "u-1", "", "o-1", OrderStatusRequest{Status: "CONFIRMED"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastOrderStatus != "confirmed" || got.Status != "confirmed" {
		t.Fatalf("status not applied: %q got=%+v", repo.lastOrderStatus, got)
	}
}

func TestSetInventoryThreshold_RejectsNegative(t *testing.T) {
	_, err := NewService(newFakeRepo(), &fakeUploader{}).SetInventoryThreshold(
		context.Background(), "u-1", "", "v-1", InventoryThresholdRequest{ReorderLevel: -1})

	appErr, ok := apperr.As(err)
	if !ok || appErr.Code != errInvalidStock.Code {
		t.Fatalf("expected invalid stock error, got %v", err)
	}
}

func TestSetInventoryThreshold_PassesLevel(t *testing.T) {
	repo := newFakeRepo()
	if _, err := NewService(repo, &fakeUploader{}).SetInventoryThreshold(
		context.Background(), "u-1", "", "v-9", InventoryThresholdRequest{ReorderLevel: 5}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastThresholdVar != "v-9" || repo.lastThresholdLevel != 5 {
		t.Fatalf("threshold not applied: %q %d", repo.lastThresholdVar, repo.lastThresholdLevel)
	}
}

func TestListInventoryMovements_NormalizesFilters(t *testing.T) {
	repo := newFakeRepo()
	if _, _, err := NewService(repo, &fakeUploader{}).ListInventoryMovements(
		context.Background(), "u-1", "", ListQuery{Type: "SALE", Search: " blue "}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastMovementsQuery.Type != "sale" || repo.lastMovementsQuery.Search != "blue" {
		t.Fatalf("filters not normalized: %+v", repo.lastMovementsQuery)
	}
}

func TestSanitizeCSVCell(t *testing.T) {
	cases := map[string]string{
		"=SUM(A1:A2)": "'=SUM(A1:A2)",
		"+1":          "'+1",
		"-1":          "'-1",
		"@cmd":        "'@cmd",
		"Blender":     "Blender",
		"":            "",
	}
	for in, want := range cases {
		if got := sanitizeCSVCell(in); got != want {
			t.Errorf("sanitizeCSVCell(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWriteExport_CSVEscapesFormulaAndSetsHeaders(t *testing.T) {
	rows := []ExportOrderRow{{
		OrderNumber: "SK-1", ProductName: "=SUM(A1:A2)", SKU: "SKU-1", Quantity: 2,
		UnitPrice: 10, Discount: 0, LineTotal: 20, Currency: "BDT",
		PaymentStatus: "paid", FulfillmentStatus: "confirmed", TrackingNumber: "T1",
	}}

	rec := httptest.NewRecorder()
	if err := writeExport(rec, "csv", ordersExportFile(rows)); err != nil {
		t.Fatalf("csv write failed: %v", err)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Order Number") || !strings.Contains(body, "'=SUM(A1:A2)") {
		t.Fatalf("unexpected csv body: %q", body)
	}
	if strings.Contains(body, ",=SUM") || strings.Contains(body, "\n=SUM") {
		t.Fatalf("formula was not neutralized: %q", body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Fatalf("unexpected content type: %q", ct)
	}
}

func TestWriteExport_XLSXRoundTrips(t *testing.T) {
	rows := []ExportInventoryRow{{Name: "Blender", SKU: "SKU-1", Price: 19.99, Quantity: 3, ReorderLevel: 5, ApprovalStatus: "approved"}}

	rec := httptest.NewRecorder()
	if err := writeExport(rec, "xlsx", inventoryExportFile(rows)); err != nil {
		t.Fatalf("xlsx write failed: %v", err)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("xlsx body was empty")
	}

	f, err := excelize.OpenReader(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("xlsx could not be opened: %v", err)
	}
	defer func() { _ = f.Close() }()

	header, err := f.GetCellValue("Inventory", "B1")
	if err != nil || header != "Product" {
		t.Fatalf("unexpected header cell: %q (%v)", header, err)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "spreadsheetml") {
		t.Fatalf("unexpected content type: %q", ct)
	}
}
