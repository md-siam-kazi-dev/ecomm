-- 004_indexes.sql
-- Indexes.
-- All secondary indexes from ARCHITECTURE.md.
-- Source: ARCHITECTURE.md (generated)

CREATE UNIQUE INDEX idx_users_email
ON users (LOWER(email));

CREATE UNIQUE INDEX idx_users_phone
ON users (phone)
WHERE phone IS NOT NULL;

CREATE INDEX idx_users_status
ON users (status);

CREATE INDEX idx_user_sessions_user_id
ON user_sessions(user_id);

CREATE INDEX idx_user_sessions_expires_at
ON user_sessions(expires_at);

CREATE INDEX idx_vendors_status
ON vendors(status);

CREATE INDEX idx_vendors_business_slug
ON vendors(business_slug);

CREATE INDEX idx_categories_parent_id
ON categories(parent_id);

CREATE INDEX idx_categories_sort_order
ON categories(sort_order);

CREATE INDEX idx_products_vendor_id
ON products(vendor_id);

CREATE INDEX idx_products_brand_id
ON products(brand_id);

CREATE INDEX idx_products_status
ON products(status);

CREATE INDEX idx_products_approval_status
ON products(approval_status);

CREATE INDEX idx_products_created_at
ON products(created_at DESC);

CREATE INDEX idx_variants_product_id
ON product_variants(product_id);

CREATE INDEX idx_variants_barcode
ON product_variants(barcode)
WHERE barcode IS NOT NULL;

CREATE INDEX idx_orders_user_id
ON orders(user_id);

CREATE INDEX idx_orders_status
ON orders(status);

CREATE INDEX idx_orders_created_at
ON orders(created_at DESC);

CREATE UNIQUE INDEX idx_reviews_unique_order_item
ON reviews(order_item_id);

CREATE INDEX idx_notifications_user_created
ON notifications(user_id, created_at DESC);

CREATE INDEX idx_audit_logs_user
ON audit_logs(user_id);

CREATE INDEX idx_audit_logs_entity
ON audit_logs(entity_type, entity_id);

CREATE INDEX idx_audit_logs_created_at
ON audit_logs(created_at DESC);
