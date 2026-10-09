-- 005_constraints.sql
-- Deferred constraints.
-- Circular FK between inventory_reservations and orders.
-- Source: ARCHITECTURE.md (generated)

ALTER TABLE inventory_reservations
ADD CONSTRAINT fk_reservations_order
FOREIGN KEY (order_id)
REFERENCES orders(id)
ON DELETE SET NULL;
