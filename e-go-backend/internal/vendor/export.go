package vendor

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// exportFile is a formatted, ready-to-write export.
type exportFile struct {
	Name    string
	Sheet   string
	Headers []string
	Rows    [][]string
}

var ordersExportHeaders = []string{
	"Order Number", "Order Date", "Product", "SKU", "Quantity",
	"Unit Price", "Discount", "Line Total", "Currency",
	"Payment Status", "Fulfilment Status", "Tracking Number",
}

// ordersExportFile maps order lines to the export grid (one row per item).
func ordersExportFile(rows []ExportOrderRow) exportFile {
	out := make([][]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, []string{
			r.OrderNumber,
			r.OrderDate.UTC().Format("2006-01-02 15:04"),
			r.ProductName,
			r.SKU,
			strconv.Itoa(r.Quantity),
			money(r.UnitPrice),
			money(r.Discount),
			money(r.LineTotal),
			r.Currency,
			r.PaymentStatus,
			r.FulfillmentStatus,
			r.TrackingNumber,
		})
	}
	return exportFile{Name: "vendor-orders", Sheet: "Orders", Headers: ordersExportHeaders, Rows: out}
}

var inventoryExportHeaders = []string{
	"Product ID", "Product", "SKU", "Price", "Cost Price", "On Hand",
	"Reserved", "Available", "Reorder Level", "Stock Status",
	"Approval Status", "Updated At",
}

// inventoryExportFile maps inventory rows to the export grid (one row per variant).
func inventoryExportFile(rows []ExportInventoryRow) exportFile {
	out := make([][]string, 0, len(rows))
	for _, r := range rows {
		cost := ""
		if r.CostPrice != nil {
			cost = money(*r.CostPrice)
		}
		updated := ""
		if r.UpdatedAt != nil {
			updated = r.UpdatedAt.UTC().Format("2006-01-02 15:04")
		}
		out = append(out, []string{
			r.ProductID,
			r.Name,
			r.SKU,
			money(r.Price),
			cost,
			strconv.FormatInt(r.Quantity, 10),
			strconv.FormatInt(r.Reserved, 10),
			strconv.FormatInt(r.Quantity-r.Reserved, 10),
			strconv.FormatInt(r.ReorderLevel, 10),
			stockStatus(r.Quantity, r.ReorderLevel),
			r.ApprovalStatus,
			updated,
		})
	}
	return exportFile{Name: "vendor-inventory", Sheet: "Inventory", Headers: inventoryExportHeaders, Rows: out}
}

func money(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }

// stockStatus mirrors the inventory list's status labelling.
func stockStatus(quantity, reorder int64) string {
	switch {
	case quantity <= 0:
		return "out"
	case reorder > 0 && quantity <= reorder:
		return "low"
	default:
		return "in"
	}
}

// writeExport streams the export as CSV (default) or XLSX.
func writeExport(w http.ResponseWriter, format string, file exportFile) error {
	if format == "xlsx" {
		return writeXLSX(w, file)
	}
	return writeCSV(w, file)
}

func writeCSV(w http.ResponseWriter, file exportFile) error {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+file.Name+`.csv"`)
	// UTF-8 BOM so Excel opens the file as UTF-8.
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}

	cw := csv.NewWriter(w)
	if err := cw.Write(file.Headers); err != nil {
		return err
	}
	for _, row := range file.Rows {
		safe := make([]string, len(row))
		for i, cell := range row {
			safe[i] = sanitizeCSVCell(cell)
		}
		if err := cw.Write(safe); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func writeXLSX(w http.ResponseWriter, file exportFile) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheet := file.Sheet
	if sheet == "" {
		sheet = "Sheet1"
	}
	if sheet != "Sheet1" {
		if err := f.SetSheetName("Sheet1", sheet); err != nil {
			return err
		}
	}

	for col, header := range file.Headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cell, header); err != nil {
			return err
		}
	}
	for i, row := range file.Rows {
		for col, value := range row {
			cell, err := excelize.CoordinatesToCellName(col+1, i+2)
			if err != nil {
				return err
			}
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				return err
			}
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="`+file.Name+`.xlsx"`)
	_, err = w.Write(buf.Bytes())
	return err
}

// sanitizeCSVCell neutralizes spreadsheet formula injection: a leading
// = + - @ (or tab/CR) is prefixed so Excel and Sheets treat the cell as text.
func sanitizeCSVCell(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + s
	}
	return s
}
