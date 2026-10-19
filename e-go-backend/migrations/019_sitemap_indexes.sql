-- 019_sitemap_indexes.sql
-- Supporting indexes for the public sitemap keyset scans
-- (GET /api/v1/sitemap/products and GET /api/v1/sitemap/stores).
--
-- Applied per-statement (no transaction wrapper), matching migrations 007-018.
-- Additive only: the endpoints work without these indexes (they just seq-scan);
-- there is no migration runner, so the operator applies this file manually.

-- Products: the sitemap scan filters on the public visibility predicate
-- (status = 'active' AND approval_status = 'approved'), then pages by
-- ORDER BY (updated_at, id) with a tuple keyset filter. This composite index
-- covers the equality prefix and keeps the ordering/keyset seek in-index.
CREATE INDEX IF NOT EXISTS idx_products_public_updated
    ON products (status, approval_status, updated_at, id);

-- Stores: the sitemap scan filters on status = 'active' and pages by
-- ORDER BY (updated_at, id), covered by the same shape.
CREATE INDEX IF NOT EXISTS idx_stores_status_updated
    ON stores (status, updated_at, id);
