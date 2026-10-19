package vendor

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/notify"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// errOrderNotFound is returned when an order has no items for the vendor.
var errOrderNotFound = apperr.New(http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")

// errOrderFinalized is returned when the order is already delivered or closed
// and can no longer receive shipment updates.
var errOrderFinalized = apperr.New(
	http.StatusConflict, "ORDER_ALREADY_FINAL",
	"This order is already delivered or closed and cannot be updated",
)

// errInvalidTransition is returned when an order status change does not move
// the order forward.
var errInvalidTransition = apperr.New(
	http.StatusUnprocessableEntity, "INVALID_TRANSITION",
	"That order status change is not allowed from the order's current status",
)

// nonRevenueOrderStatuses are excluded from revenue figures.
const nonRevenueFilter = `o.status NOT IN ('cancelled', 'returned', 'refunded')`

// OrderDetail loads the vendor-scoped order view (items, address, shipment, timeline).
func (r *Repository) OrderDetail(ctx context.Context, vendorID, orderID string) (OrderDetailRecord, error) {
	var rec OrderDetailRecord
	var hasItems bool
	err := r.db.QueryRow(ctx, `
		SELECT o.id, o.order_number, o.status::text, o.currency, o.grand_total::float8, o.created_at,
		       EXISTS (SELECT 1 FROM order_items oi WHERE oi.order_id = o.id AND oi.vendor_id = $2)
		FROM orders o
		WHERE o.id = $1
	`, orderID, vendorID).Scan(
		&rec.OrderID, &rec.OrderNumber, &rec.Status, &rec.Currency, &rec.GrandTotal, &rec.CreatedAt, &hasItems,
	)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !hasItems) {
		return OrderDetailRecord{}, errOrderNotFound
	}
	if err != nil {
		return OrderDetailRecord{}, apperr.ErrInternal.WithCause(err)
	}

	// The vendor's lines.
	rows, err := r.db.Query(ctx, `
		SELECT oi.id, COALESCE(oi.product_id::text, ''), oi.product_name, COALESCE(oi.product_sku, ''),
		       oi.quantity, oi.unit_price::float8, oi.total::float8
		FROM order_items oi
		WHERE oi.order_id = $1 AND oi.vendor_id = $2
		ORDER BY oi.created_at
	`, orderID, vendorID)
	if err != nil {
		return OrderDetailRecord{}, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()
	for rows.Next() {
		var line OrderLineRecord
		if err := rows.Scan(&line.OrderItemID, &line.ProductID, &line.ProductName, &line.ProductSKU,
			&line.Quantity, &line.UnitPrice, &line.LineTotal); err != nil {
			return OrderDetailRecord{}, apperr.ErrInternal.WithCause(err)
		}
		rec.Items = append(rec.Items, line)
		rec.VendorRevenue += line.LineTotal
		rec.VendorItemCount += int64(line.Quantity)
	}
	if err := rows.Err(); err != nil {
		return OrderDetailRecord{}, apperr.ErrInternal.WithCause(err)
	}

	// Shipping address.
	err = r.db.QueryRow(ctx, `
		SELECT full_name, COALESCE(phone, ''), address_line_1, COALESCE(address_line_2, ''),
		       city, COALESCE(state, ''), COALESCE(postal_code, ''), country
		FROM order_addresses
		WHERE order_id = $1 AND address_type = 'shipping'
		ORDER BY created_at LIMIT 1
	`, orderID).Scan(
		&rec.Address.FullName, &rec.Address.Phone, &rec.Address.Line1, &rec.Address.Line2,
		&rec.Address.City, &rec.Address.State, &rec.Address.PostalCode, &rec.Address.Country,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return OrderDetailRecord{}, apperr.ErrInternal.WithCause(err)
	}

	// Shipment (latest for this order + vendor).
	var shipment ShipmentRecord
	err = r.db.QueryRow(ctx, `
		SELECT id, COALESCE(carrier, ''), COALESCE(tracking_number, ''), status::text,
		       shipped_at, delivered_at, created_at
		FROM shipments
		WHERE order_id = $1 AND vendor_id = $2
		ORDER BY created_at DESC LIMIT 1
	`, orderID, vendorID).Scan(
		&shipment.ID, &shipment.Carrier, &shipment.TrackingNumber, &shipment.Status,
		&shipment.ShippedAt, &shipment.DeliveredAt, &shipment.CreatedAt,
	)
	if err == nil {
		rec.Shipment = &shipment

		events, err := r.db.Query(ctx, `
			SELECT id, status::text, COALESCE(location, ''), COALESCE(description, ''), event_time, created_at
			FROM tracking_events
			WHERE shipment_id = $1
			ORDER BY event_time, created_at
		`, shipment.ID)
		if err != nil {
			return OrderDetailRecord{}, apperr.ErrInternal.WithCause(err)
		}
		defer events.Close()
		for events.Next() {
			var ev TrackingEventRecord
			if err := events.Scan(&ev.ID, &ev.Status, &ev.Location, &ev.Description, &ev.EventTime, &ev.CreatedAt); err != nil {
				return OrderDetailRecord{}, apperr.ErrInternal.WithCause(err)
			}
			rec.Events = append(rec.Events, ev)
		}
		if err := events.Err(); err != nil {
			return OrderDetailRecord{}, apperr.ErrInternal.WithCause(err)
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return OrderDetailRecord{}, apperr.ErrInternal.WithCause(err)
	}

	return rec, nil
}

// UpsertShipmentEvent records a tracking update for the vendor's shipment of an
// order, creating the shipment on first use, and advances the order status.
func (r *Repository) UpsertShipmentEvent(ctx context.Context, vendorID, orderID, actorID string, p ShipmentParams) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var hasItems bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM order_items WHERE order_id = $1 AND vendor_id = $2)
	`, orderID, vendorID).Scan(&hasItems); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if !hasItems {
		return errOrderNotFound
	}

	var currentOrderStatus, buyerID, orderNumber string
	if err := tx.QueryRow(ctx, `SELECT status::text, user_id::text, order_number FROM orders WHERE id = $1`, orderID).Scan(&currentOrderStatus, &buyerID, &orderNumber); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	// Delivered/closed orders accept no further shipment updates.
	switch currentOrderStatus {
	case "delivered", "cancelled", "returned", "refunded":
		return errOrderFinalized
	}

	eventTime := time.Now()
	if p.EventTime != nil {
		eventTime = *p.EventTime
	}
	var shippedAt, deliveredAt *time.Time
	switch p.Status {
	case "shipped", "in_transit", "out_for_delivery":
		shippedAt = &eventTime
	case "delivered":
		shippedAt = &eventTime
		deliveredAt = &eventTime
	}

	var shipmentID string
	err = tx.QueryRow(ctx, `
		SELECT id FROM shipments WHERE order_id = $1 AND vendor_id = $2 ORDER BY created_at LIMIT 1
	`, orderID, vendorID).Scan(&shipmentID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		if err := tx.QueryRow(ctx, `
			INSERT INTO shipments (order_id, vendor_id, carrier, tracking_number, status, shipped_at, delivered_at)
			VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5::shipment_status, $6, $7)
			RETURNING id
		`, orderID, vendorID, p.Carrier, p.TrackingNumber, p.Status, shippedAt, deliveredAt).Scan(&shipmentID); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
		// Record the shipment's items once.
		if _, err := tx.Exec(ctx, `
			INSERT INTO shipment_items (shipment_id, order_item_id, quantity)
			SELECT $1, oi.id, oi.quantity FROM order_items oi
			WHERE oi.order_id = $2 AND oi.vendor_id = $3
		`, shipmentID, orderID, vendorID); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
	case err != nil:
		return apperr.ErrInternal.WithCause(err)
	default:
		if _, err := tx.Exec(ctx, `
			UPDATE shipments
			SET carrier = COALESCE(NULLIF($3, ''), carrier),
			    tracking_number = COALESCE(NULLIF($4, ''), tracking_number),
			    status = $5::shipment_status,
			    shipped_at = COALESCE($6, shipped_at),
			    delivered_at = COALESCE($7, delivered_at),
			    updated_at = NOW()
			WHERE id = $1 AND vendor_id = $2
		`, shipmentID, vendorID, p.Carrier, p.TrackingNumber, p.Status, shippedAt, deliveredAt); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO tracking_events (shipment_id, status, location, description, event_time)
		VALUES ($1, $2::shipment_status, NULLIF($3, ''), NULLIF($4, ''), $5)
	`, shipmentID, p.Status, p.Location, p.Description, eventTime); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	// A failed delivery attempt is a customer-facing exception.
	if p.Status == "failed" {
		if err := notify.Insert(ctx, tx, buyerID, "return_update", "Delivery attempt failed",
			fmt.Sprintf("A delivery attempt for order %s was unsuccessful.", orderNumber),
			map[string]any{"order_number": orderNumber}); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id)
		VALUES ($1, 'order.shipment_updated', 'order', $2)`, actorID, orderID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	// Advance the order status monotonically.
	if next := orderStatusForShipment(p.Status); next != "" &&
		orderStatusRank(next) > orderStatusRank(currentOrderStatus) {
		if _, err := tx.Exec(ctx, `UPDATE orders SET status = $2::order_status, updated_at = NOW() WHERE id = $1`, orderID, next); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO order_status_history (order_id, old_status, new_status, changed_by, notes)
			VALUES ($1, $2::order_status, $3::order_status, $4, 'Shipment update')
		`, orderID, currentOrderStatus, next, actorID); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
		if err := notifyOrderEvent(ctx, tx, buyerID, orderNumber, next); err != nil {
			return apperr.ErrInternal.WithCause(err)
		}
	}

	return tx.Commit(ctx)
}

// SetOrderStatus applies an order-level fulfilment transition (confirm,
// process, pack, ship, deliver) for an order the vendor has items in. The
// transition must move the order forward; terminal orders are rejected.
func (r *Repository) SetOrderStatus(ctx context.Context, vendorID, orderID, actorID, status, note string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var current, buyerID, orderNumber string
	err = tx.QueryRow(ctx, `
		SELECT o.status::text, o.user_id::text, o.order_number
		FROM orders o
		WHERE o.id = $1
		  AND EXISTS (SELECT 1 FROM order_items oi WHERE oi.order_id = o.id AND oi.vendor_id = $2)
	`, orderID, vendorID).Scan(&current, &buyerID, &orderNumber)
	if errors.Is(err, pgx.ErrNoRows) {
		return errOrderNotFound
	}
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	switch current {
	case "delivered", "cancelled", "returned", "refunded":
		return errOrderFinalized
	}

	if orderStatusRank(status) < 0 || orderStatusRank(status) <= orderStatusRank(current) {
		return errInvalidTransition
	}

	if _, err := tx.Exec(ctx, `UPDATE orders SET status = $2::order_status, updated_at = NOW() WHERE id = $1`, orderID, status); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO order_status_history (order_id, old_status, new_status, changed_by, notes)
		VALUES ($1, $2::order_status, $3::order_status, $4, NULLIF($5, ''))
	`, orderID, current, status, actorID, note); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id)
		VALUES ($1, 'order.status_changed', 'order', $2)`, actorID, orderID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if err := notifyOrderEvent(ctx, tx, buyerID, orderNumber, status); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	return tx.Commit(ctx)
}

// notifyOrderEvent sends the buyer an in-app notification for a meaningful
// order transition. Statuses without a customer-facing event are skipped.
func notifyOrderEvent(ctx context.Context, q notify.Querier, buyerID, orderNumber, status string) error {
	var ntype, title, message string
	switch status {
	case "confirmed":
		ntype, title, message = "order_confirmed", "Order confirmed", fmt.Sprintf("Your order %s has been confirmed.", orderNumber)
	case "shipped":
		ntype, title, message = "order_shipped", "Order shipped", fmt.Sprintf("Your order %s is on its way.", orderNumber)
	case "delivered":
		ntype, title, message = "order_delivered", "Order delivered", fmt.Sprintf("Your order %s was delivered.", orderNumber)
	default:
		return nil
	}
	return notify.Insert(ctx, q, buyerID, ntype, title, message, map[string]any{"order_number": orderNumber})
}

// Revenue aggregates the vendor's revenue, series, top products and statuses.
func (r *Repository) Revenue(ctx context.Context, vendorID string) (RevenueSummary, error) {
	var s RevenueSummary
	if err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(oi.total), 0)::float8,
		       COUNT(DISTINCT oi.order_id),
		       COALESCE(SUM(oi.quantity), 0)::bigint
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.vendor_id = $1 AND `+nonRevenueFilter,
		vendorID,
	).Scan(&s.Revenue, &s.OrderCount, &s.UnitsSold); err != nil {
		return RevenueSummary{}, apperr.ErrInternal.WithCause(err)
	}
	if s.OrderCount > 0 {
		s.AverageOrderValue = round2(s.Revenue / float64(s.OrderCount))
	}

	// Monthly series (last 6 months).
	byMonth := map[string]*RevenuePoint{}
	now := time.Now()
	s.Series = make([]RevenuePoint, 0, 6)
	for offset := 5; offset >= 0; offset-- {
		month := now.AddDate(0, -offset, 0)
		point := RevenuePoint{Month: month.Format("2006-01")}
		s.Series = append(s.Series, point)
		byMonth[point.Month] = &s.Series[len(s.Series)-1]
	}
	rows, err := r.db.Query(ctx, `
		SELECT to_char(date_trunc('month', o.created_at), 'YYYY-MM'),
		       COALESCE(SUM(oi.total), 0)::float8,
		       COUNT(DISTINCT oi.order_id)
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.vendor_id = $1 AND `+nonRevenueFilter+`
		  AND o.created_at >= date_trunc('month', NOW()) - interval '5 months'
		GROUP BY 1
	`, vendorID)
	if err != nil {
		return RevenueSummary{}, apperr.ErrInternal.WithCause(err)
	}
	for rows.Next() {
		var (
			month   string
			revenue float64
			orders  int64
		)
		if err := rows.Scan(&month, &revenue, &orders); err != nil {
			rows.Close()
			return RevenueSummary{}, apperr.ErrInternal.WithCause(err)
		}
		if point, ok := byMonth[month]; ok {
			point.Revenue = revenue
			point.Orders = orders
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return RevenueSummary{}, apperr.ErrInternal.WithCause(err)
	}

	// Top products by revenue.
	topRows, err := r.db.Query(ctx, `
		SELECT COALESCE(oi.product_id::text, ''), COALESCE(p.name, oi.product_name),
		       COALESCE(SUM(oi.quantity), 0)::bigint, COALESCE(SUM(oi.total), 0)::float8
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		LEFT JOIN products p ON p.id = oi.product_id
		WHERE oi.vendor_id = $1 AND `+nonRevenueFilter+`
		GROUP BY 1, 2
		ORDER BY 4 DESC
		LIMIT 5
	`, vendorID)
	if err != nil {
		return RevenueSummary{}, apperr.ErrInternal.WithCause(err)
	}
	defer topRows.Close()
	s.TopProducts = make([]TopProduct, 0)
	for topRows.Next() {
		var p TopProduct
		if err := topRows.Scan(&p.ProductID, &p.Name, &p.UnitsSold, &p.Revenue); err != nil {
			return RevenueSummary{}, apperr.ErrInternal.WithCause(err)
		}
		s.TopProducts = append(s.TopProducts, p)
	}
	if err := topRows.Err(); err != nil {
		return RevenueSummary{}, apperr.ErrInternal.WithCause(err)
	}

	// Orders by status (revenue excluded for non-revenue statuses is not needed
	// here — this is a lifecycle count).
	statusRows, err := r.db.Query(ctx, `
		SELECT o.status::text, COUNT(DISTINCT o.id), COALESCE(SUM(oi.total), 0)::float8
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.vendor_id = $1
		GROUP BY 1
		ORDER BY 1
	`, vendorID)
	if err != nil {
		return RevenueSummary{}, apperr.ErrInternal.WithCause(err)
	}
	defer statusRows.Close()
	s.Statuses = make([]StatusCount, 0)
	for statusRows.Next() {
		var sc StatusCount
		if err := statusRows.Scan(&sc.Status, &sc.Count, &sc.Revenue); err != nil {
			return RevenueSummary{}, apperr.ErrInternal.WithCause(err)
		}
		s.Statuses = append(s.Statuses, sc)
	}
	if err := statusRows.Err(); err != nil {
		return RevenueSummary{}, apperr.ErrInternal.WithCause(err)
	}

	return s, nil
}

// orderStatusForShipment maps a shipment status onto an order status ("" = no change).
func orderStatusForShipment(shipmentStatus string) string {
	switch shipmentStatus {
	case "packed":
		return "packed"
	case "shipped", "in_transit", "out_for_delivery":
		return "shipped"
	case "delivered":
		return "delivered"
	default:
		return ""
	}
}

func orderStatusRank(status string) int {
	switch status {
	case "pending":
		return 0
	case "confirmed":
		return 1
	case "processing":
		return 2
	case "packed":
		return 3
	case "shipped":
		return 4
	case "delivered":
		return 5
	default:
		return -1
	}
}
