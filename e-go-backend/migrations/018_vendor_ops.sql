-- 018_vendor_ops.sql
-- Vendor operations: product review metadata, inventory movement history
-- (previous -> new stock), and a supporting index.
--
-- Applied per-statement (no transaction wrapper), matching migrations 007-017.

-- Review metadata for every admin decision (approve keeps approved_at too).
ALTER TABLE products ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ;
ALTER TABLE products ADD COLUMN IF NOT EXISTS reviewed_by UUID;

-- Movement history shows the stock level before and after a change.
ALTER TABLE inventory_movements ADD COLUMN IF NOT EXISTS quantity_before INTEGER;
ALTER TABLE inventory_movements ADD COLUMN IF NOT EXISTS quantity_after INTEGER;

CREATE INDEX IF NOT EXISTS idx_inventory_movements_variant_created
    ON inventory_movements (variant_id, created_at DESC);
