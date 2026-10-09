-- 003_tables.sql
-- Tables (51 + vendor_ledger).
-- Created in dependency order from ARCHITECTURE.md.
-- Source: ARCHITECTURE.md (generated)

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    email VARCHAR(255) NOT NULL,
    phone VARCHAR(30),

    password_hash TEXT NOT NULL,

    status user_status NOT NULL DEFAULT 'pending',

    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    phone_verified BOOLEAN NOT NULL DEFAULT FALSE,

    last_login_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL UNIQUE,

    first_name VARCHAR(100),
    last_name VARCHAR(100),

    avatar_url TEXT,

    date_of_birth DATE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user_profiles_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL,

    refresh_token_hash TEXT NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,

    revoked_at TIMESTAMPTZ,

    ip_address INET,

    user_agent TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_sessions_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE user_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL,

    label VARCHAR(50),

    full_name VARCHAR(200) NOT NULL,

    phone VARCHAR(30) NOT NULL,

    address_line_1 VARCHAR(255) NOT NULL,
    address_line_2 VARCHAR(255),

    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    postal_code VARCHAR(30),

    country VARCHAR(100) NOT NULL DEFAULT 'Bangladesh',

    is_default BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_addresses_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name user_role_type NOT NULL UNIQUE,

    description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(100) NOT NULL UNIQUE,

    description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE role_permissions (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,

    PRIMARY KEY (role_id, permission_id),

    CONSTRAINT fk_role_permissions_role
        FOREIGN KEY (role_id)
        REFERENCES roles(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_role_permissions_permission
        FOREIGN KEY (permission_id)
        REFERENCES permissions(id)
        ON DELETE CASCADE
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,

    PRIMARY KEY (user_id, role_id),

    CONSTRAINT fk_user_roles_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_user_roles_role
        FOREIGN KEY (role_id)
        REFERENCES roles(id)
        ON DELETE CASCADE
);

CREATE TABLE vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL UNIQUE,

    business_name VARCHAR(255) NOT NULL,
    business_slug VARCHAR(255) NOT NULL UNIQUE,

    business_email VARCHAR(255),
    business_phone VARCHAR(30),

    description TEXT,

    logo_url TEXT,

    status vendor_status NOT NULL DEFAULT 'pending',

    approved_by UUID,
    approved_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_vendors_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_vendors_approved_by
        FOREIGN KEY (approved_by)
        REFERENCES users(id)
        ON DELETE SET NULL
);

CREATE TABLE vendor_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    vendor_id UUID NOT NULL,

    document_type VARCHAR(100) NOT NULL,

    document_url TEXT NOT NULL,

    status approval_status NOT NULL DEFAULT 'pending',

    reviewed_by UUID,

    reviewed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_vendor_documents_vendor
        FOREIGN KEY (vendor_id)
        REFERENCES vendors(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_vendor_documents_reviewer
        FOREIGN KEY (reviewed_by)
        REFERENCES users(id)
        ON DELETE SET NULL
);

CREATE TABLE brands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(150) NOT NULL,

    slug VARCHAR(180) NOT NULL UNIQUE,

    description TEXT,

    logo_url TEXT,

    status BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    parent_id UUID,

    name VARCHAR(150) NOT NULL,

    slug VARCHAR(180) NOT NULL UNIQUE,

    description TEXT,

    image_url TEXT,

    sort_order INTEGER NOT NULL DEFAULT 0,

    status BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_categories_parent
        FOREIGN KEY (parent_id)
        REFERENCES categories(id)
        ON DELETE SET NULL
);

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    vendor_id UUID NOT NULL,

    brand_id UUID,

    name VARCHAR(255) NOT NULL,

    slug VARCHAR(300) NOT NULL UNIQUE,

    description TEXT,

    short_description TEXT,

    status product_status NOT NULL DEFAULT 'draft',

    approval_status approval_status NOT NULL DEFAULT 'pending',

    seo_title VARCHAR(255),
    seo_description TEXT,

    approved_by UUID,
    approved_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_products_vendor
        FOREIGN KEY (vendor_id)
        REFERENCES vendors(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_products_brand
        FOREIGN KEY (brand_id)
        REFERENCES brands(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_products_approved_by
        FOREIGN KEY (approved_by)
        REFERENCES users(id)
        ON DELETE SET NULL
);

CREATE TABLE product_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    product_id UUID NOT NULL,

    sku VARCHAR(100) NOT NULL UNIQUE,

    barcode VARCHAR(100),

    price NUMERIC(19,4) NOT NULL,
    compare_at_price NUMERIC(19,4),

    cost_price NUMERIC(19,4),

    weight_grams NUMERIC(12,3),

    status BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_variants_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_variant_price
        CHECK (price >= 0),

    CONSTRAINT chk_variant_compare_price
        CHECK (
            compare_at_price IS NULL
            OR compare_at_price >= 0
        ),

    CONSTRAINT chk_variant_cost_price
        CHECK (
            cost_price IS NULL
            OR cost_price >= 0
        )
);

CREATE TABLE product_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    product_id UUID NOT NULL,

    variant_id UUID,

    url TEXT NOT NULL,

    alt_text VARCHAR(255),

    sort_order INTEGER NOT NULL DEFAULT 0,

    is_primary BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_product_images_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_product_images_variant
        FOREIGN KEY (variant_id)
        REFERENCES product_variants(id)
        ON DELETE CASCADE
);

CREATE TABLE product_categories (
    product_id UUID NOT NULL,
    category_id UUID NOT NULL,

    PRIMARY KEY (product_id, category_id),

    CONSTRAINT fk_product_categories_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_product_categories_category
        FOREIGN KEY (category_id)
        REFERENCES categories(id)
        ON DELETE CASCADE
);

CREATE TABLE product_attributes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(100) NOT NULL UNIQUE,

    display_name VARCHAR(150) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE attribute_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    attribute_id UUID NOT NULL,

    value VARCHAR(150) NOT NULL,

    display_value VARCHAR(150),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_attribute_values_attribute
        FOREIGN KEY (attribute_id)
        REFERENCES product_attributes(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_attribute_value
        UNIQUE(attribute_id, value)
);

CREATE TABLE variant_attribute_values (
    variant_id UUID NOT NULL,

    attribute_value_id UUID NOT NULL,

    PRIMARY KEY (variant_id, attribute_value_id),

    CONSTRAINT fk_variant_attribute_variant
        FOREIGN KEY (variant_id)
        REFERENCES product_variants(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_variant_attribute_value
        FOREIGN KEY (attribute_value_id)
        REFERENCES attribute_values(id)
        ON DELETE CASCADE
);

CREATE TABLE inventory_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    vendor_id UUID,

    name VARCHAR(150) NOT NULL,

    address_line_1 VARCHAR(255),
    address_line_2 VARCHAR(255),

    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(30),
    country VARCHAR(100),

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_inventory_locations_vendor
        FOREIGN KEY (vendor_id)
        REFERENCES vendors(id)
        ON DELETE CASCADE
);

CREATE TABLE inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    variant_id UUID NOT NULL,

    location_id UUID NOT NULL,

    quantity INTEGER NOT NULL DEFAULT 0,

    reserved_quantity INTEGER NOT NULL DEFAULT 0,

    reorder_level INTEGER NOT NULL DEFAULT 0,

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_inventory_variant
        FOREIGN KEY (variant_id)
        REFERENCES product_variants(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_inventory_location
        FOREIGN KEY (location_id)
        REFERENCES inventory_locations(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_inventory_variant_location
        UNIQUE(variant_id, location_id),

    CONSTRAINT chk_inventory_quantity
        CHECK (quantity >= 0),

    CONSTRAINT chk_inventory_reserved
        CHECK (
            reserved_quantity >= 0
            AND reserved_quantity <= quantity
        )
);

CREATE TABLE inventory_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    variant_id UUID NOT NULL,

    location_id UUID NOT NULL,

    movement_type inventory_movement_type NOT NULL,

    quantity INTEGER NOT NULL,

    reference_type VARCHAR(50),

    reference_id UUID,

    notes TEXT,

    created_by UUID,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_inventory_movements_variant
        FOREIGN KEY (variant_id)
        REFERENCES product_variants(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_inventory_movements_location
        FOREIGN KEY (location_id)
        REFERENCES inventory_locations(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_inventory_movements_user
        FOREIGN KEY (created_by)
        REFERENCES users(id)
        ON DELETE SET NULL,

    CONSTRAINT chk_inventory_movement_quantity
        CHECK (quantity > 0)
);

CREATE TABLE inventory_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    variant_id UUID NOT NULL,

    location_id UUID NOT NULL,

    order_id UUID,

    quantity INTEGER NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,

    released_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_reservations_variant
        FOREIGN KEY (variant_id)
        REFERENCES product_variants(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_reservations_location
        FOREIGN KEY (location_id)
        REFERENCES inventory_locations(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_reservation_quantity
        CHECK (quantity > 0)
);

CREATE TABLE carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID,

    session_id VARCHAR(255),

    status cart_status NOT NULL DEFAULT 'active',

    currency CHAR(3) NOT NULL DEFAULT 'BDT',

    expires_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_carts_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    cart_id UUID NOT NULL,

    variant_id UUID NOT NULL,

    quantity INTEGER NOT NULL,

    unit_price NUMERIC(19,4) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_cart_items_cart
        FOREIGN KEY (cart_id)
        REFERENCES carts(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_cart_items_variant
        FOREIGN KEY (variant_id)
        REFERENCES product_variants(id)
        ON DELETE RESTRICT,

    CONSTRAINT uq_cart_variant
        UNIQUE(cart_id, variant_id),

    CONSTRAINT chk_cart_quantity
        CHECK (quantity > 0),

    CONSTRAINT chk_cart_unit_price
        CHECK (unit_price >= 0)
);

CREATE TABLE wishlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL UNIQUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_wishlists_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE wishlist_items (
    wishlist_id UUID NOT NULL,

    product_id UUID NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (wishlist_id, product_id),

    CONSTRAINT fk_wishlist_items_wishlist
        FOREIGN KEY (wishlist_id)
        REFERENCES wishlists(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_wishlist_items_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE CASCADE
);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    order_number VARCHAR(50) NOT NULL UNIQUE,

    user_id UUID NOT NULL,

    status order_status NOT NULL DEFAULT 'pending',

    currency CHAR(3) NOT NULL DEFAULT 'BDT',

    subtotal NUMERIC(19,4) NOT NULL DEFAULT 0,
    discount_total NUMERIC(19,4) NOT NULL DEFAULT 0,
    tax_total NUMERIC(19,4) NOT NULL DEFAULT 0,
    shipping_total NUMERIC(19,4) NOT NULL DEFAULT 0,

    grand_total NUMERIC(19,4) NOT NULL DEFAULT 0,

    coupon_code VARCHAR(100),

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_orders_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_order_amounts
        CHECK (
            subtotal >= 0
            AND discount_total >= 0
            AND tax_total >= 0
            AND shipping_total >= 0
            AND grand_total >= 0
        )
);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    order_id UUID NOT NULL,

    vendor_id UUID NOT NULL,

    product_id UUID,

    variant_id UUID,

    product_name VARCHAR(255) NOT NULL,

    product_sku VARCHAR(100),

    quantity INTEGER NOT NULL,

    unit_price NUMERIC(19,4) NOT NULL,

    discount NUMERIC(19,4) NOT NULL DEFAULT 0,

    tax NUMERIC(19,4) NOT NULL DEFAULT 0,

    cost_price NUMERIC(19,4),

    subtotal NUMERIC(19,4) NOT NULL,

    total NUMERIC(19,4) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_order_items_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_order_items_vendor
        FOREIGN KEY (vendor_id)
        REFERENCES vendors(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_order_items_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_order_items_variant
        FOREIGN KEY (variant_id)
        REFERENCES product_variants(id)
        ON DELETE SET NULL,

    CONSTRAINT chk_order_item_quantity
        CHECK (quantity > 0),

    CONSTRAINT chk_order_item_amounts
        CHECK (
            unit_price >= 0
            AND discount >= 0
            AND tax >= 0
            AND subtotal >= 0
            AND total >= 0
        )
);

CREATE TABLE order_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    order_id UUID NOT NULL,

    address_type VARCHAR(20) NOT NULL,

    full_name VARCHAR(200) NOT NULL,

    phone VARCHAR(30) NOT NULL,

    address_line_1 VARCHAR(255) NOT NULL,
    address_line_2 VARCHAR(255),

    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    postal_code VARCHAR(30),
    country VARCHAR(100) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_order_addresses_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE CASCADE
);

CREATE TABLE order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    order_id UUID NOT NULL,

    old_status order_status,

    new_status order_status NOT NULL,

    changed_by UUID,

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_order_history_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_order_history_user
        FOREIGN KEY (changed_by)
        REFERENCES users(id)
        ON DELETE SET NULL
);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    order_id UUID NOT NULL,

    provider VARCHAR(100) NOT NULL,

    provider_payment_id VARCHAR(255),

    status payment_status NOT NULL DEFAULT 'pending',

    amount NUMERIC(19,4) NOT NULL,

    currency CHAR(3) NOT NULL DEFAULT 'BDT',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_payments_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_payment_amount
        CHECK (amount >= 0)
);

CREATE TABLE payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    payment_id UUID NOT NULL,

    transaction_type payment_transaction_type NOT NULL,

    amount NUMERIC(19,4) NOT NULL,

    provider_transaction_id VARCHAR(255),

    status payment_status NOT NULL,

    metadata JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_payment_transactions_payment
        FOREIGN KEY (payment_id)
        REFERENCES payments(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_payment_transaction_amount
        CHECK (amount >= 0)
);

CREATE TABLE refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    payment_id UUID NOT NULL,

    order_id UUID NOT NULL,

    amount NUMERIC(19,4) NOT NULL,

    reason TEXT,

    status payment_status NOT NULL DEFAULT 'pending',

    provider_refund_id VARCHAR(255),

    processed_by UUID,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_refunds_payment
        FOREIGN KEY (payment_id)
        REFERENCES payments(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_refunds_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_refunds_user
        FOREIGN KEY (processed_by)
        REFERENCES users(id)
        ON DELETE SET NULL,

    CONSTRAINT chk_refund_amount
        CHECK (amount > 0)
);

CREATE TABLE shipping_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(150) NOT NULL,

    description TEXT,

    price NUMERIC(19,4) NOT NULL DEFAULT 0,

    estimated_min_days INTEGER,
    estimated_max_days INTEGER,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_shipping_price
        CHECK (price >= 0),

    CONSTRAINT chk_shipping_days
        CHECK (
            estimated_min_days IS NULL
            OR estimated_min_days >= 0
        )
);

CREATE TABLE shipments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    order_id UUID NOT NULL,

    vendor_id UUID NOT NULL,

    shipping_method_id UUID,

    carrier VARCHAR(150),

    tracking_number VARCHAR(255),

    status shipment_status NOT NULL DEFAULT 'pending',

    shipped_at TIMESTAMPTZ,

    delivered_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_shipments_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_shipments_vendor
        FOREIGN KEY (vendor_id)
        REFERENCES vendors(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_shipments_shipping_method
        FOREIGN KEY (shipping_method_id)
        REFERENCES shipping_methods(id)
        ON DELETE SET NULL
);

CREATE TABLE shipment_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    shipment_id UUID NOT NULL,

    order_item_id UUID NOT NULL,

    quantity INTEGER NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_shipment_items_shipment
        FOREIGN KEY (shipment_id)
        REFERENCES shipments(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_shipment_items_order_item
        FOREIGN KEY (order_item_id)
        REFERENCES order_items(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_shipment_quantity
        CHECK (quantity > 0)
);

CREATE TABLE tracking_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    shipment_id UUID NOT NULL,

    status shipment_status NOT NULL,

    location VARCHAR(255),

    description TEXT,

    event_time TIMESTAMPTZ NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_tracking_events_shipment
        FOREIGN KEY (shipment_id)
        REFERENCES shipments(id)
        ON DELETE CASCADE
);

CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    product_id UUID NOT NULL,

    user_id UUID NOT NULL,

    order_item_id UUID NOT NULL,

    rating SMALLINT NOT NULL,

    title VARCHAR(255),

    content TEXT NOT NULL,

    status review_status NOT NULL DEFAULT 'pending',

    vendor_reply TEXT,

    vendor_replied_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_reviews_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_reviews_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_reviews_order_item
        FOREIGN KEY (order_item_id)
        REFERENCES order_items(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_review_rating
        CHECK (rating BETWEEN 1 AND 5)
);

CREATE TABLE review_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    review_id UUID NOT NULL,

    image_url TEXT NOT NULL,

    sort_order INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_review_images_review
        FOREIGN KEY (review_id)
        REFERENCES reviews(id)
        ON DELETE CASCADE
);

CREATE TABLE review_votes (
    review_id UUID NOT NULL,

    user_id UUID NOT NULL,

    is_helpful BOOLEAN NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (review_id, user_id),

    CONSTRAINT fk_review_votes_review
        FOREIGN KEY (review_id)
        REFERENCES reviews(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_review_votes_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE product_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    product_id UUID NOT NULL,

    user_id UUID NOT NULL,

    question TEXT NOT NULL,

    is_published BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_product_questions_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_product_questions_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE product_question_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    question_id UUID NOT NULL,

    user_id UUID NOT NULL,

    answer TEXT NOT NULL,

    is_official_vendor_answer BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_question_answers_question
        FOREIGN KEY (question_id)
        REFERENCES product_questions(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_question_answers_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    code VARCHAR(100) NOT NULL UNIQUE,

    description TEXT,

    type coupon_type NOT NULL,

    value NUMERIC(19,4) NOT NULL,

    minimum_order_value NUMERIC(19,4),

    maximum_discount NUMERIC(19,4),

    usage_limit INTEGER,

    per_user_limit INTEGER,

    used_count INTEGER NOT NULL DEFAULT 0,

    starts_at TIMESTAMPTZ NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_coupon_value
        CHECK (value >= 0),

    CONSTRAINT chk_coupon_dates
        CHECK (expires_at > starts_at)
);

CREATE TABLE coupon_products (
    coupon_id UUID NOT NULL,

    product_id UUID NOT NULL,

    PRIMARY KEY (coupon_id, product_id),

    CONSTRAINT fk_coupon_products_coupon
        FOREIGN KEY (coupon_id)
        REFERENCES coupons(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_coupon_products_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE CASCADE
);

CREATE TABLE coupon_categories (
    coupon_id UUID NOT NULL,

    category_id UUID NOT NULL,

    PRIMARY KEY (coupon_id, category_id),

    CONSTRAINT fk_coupon_categories_coupon
        FOREIGN KEY (coupon_id)
        REFERENCES coupons(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_coupon_categories_category
        FOREIGN KEY (category_id)
        REFERENCES categories(id)
        ON DELETE CASCADE
);

CREATE TABLE returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    order_id UUID NOT NULL,

    user_id UUID NOT NULL,

    status return_status NOT NULL DEFAULT 'requested',

    reason TEXT NOT NULL,

    admin_notes TEXT,

    reviewed_by UUID,

    reviewed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_returns_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_returns_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_returns_reviewer
        FOREIGN KEY (reviewed_by)
        REFERENCES users(id)
        ON DELETE SET NULL
);

CREATE TABLE return_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    return_id UUID NOT NULL,

    order_item_id UUID NOT NULL,

    quantity INTEGER NOT NULL,

    reason TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_return_items_return
        FOREIGN KEY (return_id)
        REFERENCES returns(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_return_items_order_item
        FOREIGN KEY (order_item_id)
        REFERENCES order_items(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_return_quantity
        CHECK (quantity > 0)
);

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL,

    type notification_type NOT NULL,

    channel notification_channel NOT NULL DEFAULT 'in_app',

    title VARCHAR(255) NOT NULL,

    message TEXT NOT NULL,

    data JSONB,

    read_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_notifications_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL UNIQUE,

    order_updates BOOLEAN NOT NULL DEFAULT TRUE,

    payment_updates BOOLEAN NOT NULL DEFAULT TRUE,

    promotional_updates BOOLEAN NOT NULL DEFAULT TRUE,

    review_updates BOOLEAN NOT NULL DEFAULT TRUE,

    email_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    sms_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    push_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_notification_preferences_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID,

    action VARCHAR(100) NOT NULL,

    entity_type VARCHAR(100) NOT NULL,

    entity_id UUID,

    old_values JSONB,

    new_values JSONB,

    ip_address INET,

    user_agent TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_audit_logs_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE SET NULL
);

CREATE TABLE vendor_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    vendor_id UUID NOT NULL,

    order_id UUID,

    order_item_id UUID,

    transaction_type VARCHAR(50) NOT NULL,

    amount NUMERIC(19,4) NOT NULL,

    currency CHAR(3) NOT NULL DEFAULT 'BDT',

    description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_vendor_ledger_vendor
        FOREIGN KEY (vendor_id)
        REFERENCES vendors(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_vendor_ledger_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_vendor_ledger_order_item
        FOREIGN KEY (order_item_id)
        REFERENCES order_items(id)
        ON DELETE SET NULL
);
