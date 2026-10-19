package vendor

import (
	"context"
	"strconv"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// maxExportRows bounds an export so one request cannot load the entire order
// history or catalog into memory.
const maxExportRows = 10000

// ExportOrders returns one row per order item for the vendor, newest first.
// Order-level fields (payment/fulfilment status, tracking) repeat on each item
// row of that order; the item-level totals are never duplicated.
func (r *Repository) ExportOrders(ctx context.Context, vendorID string, q ExportQuery) ([]ExportOrderRow, error) {
	args := []any{vendorID}
	where := " WHERE oi.vendor_id = $1"
	if q.Status != "" {
		args = append(args, q.Status)
		where += " AND o.status = $" + strconv.Itoa(len(args)) + "::order_status"
	}
	if !q.From.IsZero() {
		args = append(args, q.From)
		where += " AND o.created_at >= $" + strconv.Itoa(len(args))
	}
	if !q.To.IsZero() {
		args = append(args, q.To)
		where += " AND o.created_at < $" + strconv.Itoa(len(args))
	}
	args = append(args, maxExportRows)

	rows, err := r.db.Query(ctx, `
		SELECT o.order_number, o.created_at, oi.product_name, COALESCE(oi.product_sku, ''),
		       oi.quantity, oi.unit_price::float8, oi.discount::float8, oi.total::float8,
		       COALESCE((SELECT p.status::text FROM payments p WHERE p.order_id = o.id ORDER BY p.created_at DESC LIMIT 1), ''),
		       o.status::text,
		       COALESCE((SELECT s.tracking_number FROM shipments s
		                 WHERE s.order_id = o.id AND s.vendor_id = $1
		                 ORDER BY s.created_at DESC LIMIT 1), ''),
		       o.currency
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id`+where+`
		ORDER BY o.created_at DESC, o.id, oi.created_at
		LIMIT $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	out := make([]ExportOrderRow, 0)
	for rows.Next() {
		var row ExportOrderRow
		if err := rows.Scan(
			&row.OrderNumber, &row.OrderDate, &row.ProductName, &row.SKU,
			&row.Quantity, &row.UnitPrice, &row.Discount, &row.LineTotal,
			&row.PaymentStatus, &row.FulfillmentStatus, &row.TrackingNumber, &row.Currency,
		); err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return out, nil
}

// ExportInventory returns one row per variant for the vendor, honouring the
// same status/search filters as the inventory list.
func (r *Repository) ExportInventory(ctx context.Context, vendorID string, q ListQuery) ([]ExportInventoryRow, error) {
	where, having, args := buildInventoryFilter(vendorID, q)
	args = append(args, maxExportRows)

	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.name, v.sku, v.price::float8, v.cost_price,
		       COALESCE(SUM(i.quantity), 0)::bigint,
		       COALESCE(SUM(i.reserved_quantity), 0)::bigint,
		       COALESCE(MAX(i.reorder_level), 0)::bigint,
		       p.status::text, p.approval_status::text, MAX(i.updated_at)
		FROM product_variants v
		JOIN products p ON p.id = v.product_id
		LEFT JOIN inventory i ON i.variant_id = v.id`+where+`
		GROUP BY v.id, p.id`+having+`
		ORDER BY p.name, v.sku
		LIMIT $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	out := make([]ExportInventoryRow, 0)
	for rows.Next() {
		var row ExportInventoryRow
		if err := rows.Scan(
			&row.ProductID, &row.Name, &row.SKU, &row.Price, &row.CostPrice,
			&row.Quantity, &row.Reserved, &row.ReorderLevel,
			&row.ProductStatus, &row.ApprovalStatus, &row.UpdatedAt,
		); err != nil {
			return nil, apperr.ErrInternal.WithCause(err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return out, nil
}
