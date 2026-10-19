# Production E-commerce Database — ERD & Schema Specification

## 1. Database Overview

### Database

```text
PostgreSQL
Provider: Neon
```

### Primary key strategy

Use UUIDs for externally visible entity IDs.

Recommended:

```sql
id UUID PRIMARY KEY DEFAULT gen_random_uuid()
```

If PostgreSQL/Neon configuration supports UUIDv7 generation, prefer UUIDv7 for high-volume tables because it provides better index locality.

### Money

Never use `FLOAT` or `REAL` for money.

Use:

```sql
NUMERIC(19,4)
```

All monetary columns should have:

```sql
CHECK (amount >= 0)
```

where appropriate.

### Timestamps

Use:

```sql
TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Always store timestamps in UTC.

---

# 2. Database Domains

```text
01. Authentication & Users
02. Vendors
03. Catalog
04. Inventory
05. Shopping
06. Orders
07. Payments
08. Shipping
09. Reviews & Questions
10. Promotions
11. Returns
12. Notifications
13. Auditing
```

Total tables:

```text
48
```

---

# 3. Entity Relationship Overview

```text
users
 │
 ├──────── user_profiles
 ├──────── user_sessions
 ├──────── user_addresses
 ├──────── user_roles ───── roles ───── role_permissions ───── permissions
 │
 └──────── vendors
              │
              ├──────── vendor_documents
              │
              └──────── products
                         │
                         ├──────── product_variants
                         │              │
                         │              └──────── inventory
                         │
                         ├──────── product_images
                         ├──────── product_categories ───── categories
                         ├──────── product_attributes
                         └──────── product_questions
                                        │
                                        └──── product_question_answers

users
 │
 ├──── carts ───── cart_items ───── product_variants
 │
 ├──── wishlists ───── wishlist_items ───── products
 │
 └──── orders
          │
          ├──── order_items ───── products / product_variants
          ├──── order_addresses
          ├──── order_status_history
          ├──── payments
          │       ├──── payment_transactions
          │       └──── refunds
          │
          ├──── shipments
          │       ├──── shipment_items
          │       └──── tracking_events
          │
          └──── returns
                  └──── return_items

products
 │
 ├──── reviews
 │       ├──── review_images
 │       └──── review_votes
 │
 └──── coupons
         ├──── coupon_products
         └──── coupon_categories

users
 │
 └──── notifications
```

---

# 4. ENUM TYPES

Create PostgreSQL enums before creating tables.

## user_status

```sql
CREATE TYPE user_status AS ENUM (
    'pending',
    'active',
    'suspended',
    'blocked'
);
```

## user_role_type

```sql
CREATE TYPE user_role_type AS ENUM (
    'user',
    'vendor',
    'admin'
);
```

## vendor_status

```sql
CREATE TYPE vendor_status AS ENUM (
    'pending',
    'approved',
    'rejected',
    'suspended'
);
```

## product_status

```sql
CREATE TYPE product_status AS ENUM (
    'draft',
    'active',
    'inactive',
    'archived',
    'suspended'
);
```

## approval_status

```sql
CREATE TYPE approval_status AS ENUM (
    'pending',
    'approved',
    'rejected'
);
```

## inventory_movement_type

```sql
CREATE TYPE inventory_movement_type AS ENUM (
    'purchase',
    'sale',
    'reservation',
    'reservation_release',
    'return',
    'damage',
    'adjustment',
    'restock'
);
```

## cart_status

```sql
CREATE TYPE cart_status AS ENUM (
    'active',
    'converted',
    'abandoned',
    'expired'
);
```

## order_status

```sql
CREATE TYPE order_status AS ENUM (
    'pending',
    'confirmed',
    'processing',
    'packed',
    'shipped',
    'delivered',
    'cancelled',
    'returned',
    'refunded'
);
```

## payment_status

```sql
CREATE TYPE payment_status AS ENUM (
    'pending',
    'processing',
    'paid',
    'failed',
    'cancelled',
    'partially_refunded',
    'refunded'
);
```

## payment_transaction_type

```sql
CREATE TYPE payment_transaction_type AS ENUM (
    'authorization',
    'capture',
    'payment',
    'refund',
    'partial_refund',
    'void'
);
```

## shipment_status

```sql
CREATE TYPE shipment_status AS ENUM (
    'pending',
    'packed',
    'shipped',
    'in_transit',
    'out_for_delivery',
    'delivered',
    'failed',
    'returned'
);
```

## review_status

```sql
CREATE TYPE review_status AS ENUM (
    'pending',
    'published',
    'rejected',
    'hidden'
);
```

## coupon_type

```sql
CREATE TYPE coupon_type AS ENUM (
    'percentage',
    'fixed_amount',
    'free_shipping'
);
```

## return_status

```sql
CREATE TYPE return_status AS ENUM (
    'requested',
    'approved',
    'rejected',
    'in_transit',
    'received',
    'inspecting',
    'completed',
    'cancelled'
);
```

## notification_type

```sql
CREATE TYPE notification_type AS ENUM (
    'order_created',
    'order_confirmed',
    'order_shipped',
    'order_delivered',
    'payment_success',
    'payment_failed',
    'refund_created',
    'vendor_approved',
    'vendor_rejected',
    'product_approved',
    'product_rejected',
    'review_reply',
    'return_update',
    'system'
);
```

## notification_channel

```sql
CREATE TYPE notification_channel AS ENUM (
    'in_app',
    'email',
    'sms'
);
```

---

# 5. TABLE 01 — users

Central authentication/user identity table.

```sql
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
```

### Indexes

```sql
CREATE UNIQUE INDEX idx_users_email
ON users (LOWER(email));

CREATE UNIQUE INDEX idx_users_phone
ON users (phone)
WHERE phone IS NOT NULL;

CREATE INDEX idx_users_status
ON users (status);
```

---

# 6. TABLE 02 — user_profiles

```sql
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
```

---

# 7. TABLE 03 — user_sessions

Used for refresh-token/session management.

```sql
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
```

### Indexes

```sql
CREATE INDEX idx_user_sessions_user_id
ON user_sessions(user_id);

CREATE INDEX idx_user_sessions_expires_at
ON user_sessions(expires_at);
```

---

# 8. TABLE 04 — user_addresses

```sql
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
```

---

# 9. TABLE 05 — roles

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name user_role_type NOT NULL UNIQUE,

    description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Seed:

```text
user
vendor
admin
```

---

# 10. TABLE 06 — permissions

```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(100) NOT NULL UNIQUE,

    description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Examples:

```text
product.create
product.update
product.delete
product.approve

order.view
order.ship

vendor.approve

analytics.view
```

---

# 11. TABLE 07 — role_permissions

```sql
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
```

---

# 12. TABLE 08 — user_roles

```sql
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
```

A user can technically have multiple roles, but the application should enforce the marketplace business rules.

---

# 13. TABLE 09 — vendors

```sql
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
```

### Indexes

```sql
CREATE INDEX idx_vendors_status
ON vendors(status);

CREATE INDEX idx_vendors_business_slug
ON vendors(business_slug);
```

---

# 14. TABLE 10 — vendor_documents

```sql
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
```

---

# 15. TABLE 11 — brands

```sql
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
```

---

# 16. TABLE 12 — categories

Supports nested categories.

```sql
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
```

### Indexes

```sql
CREATE INDEX idx_categories_parent_id
ON categories(parent_id);

CREATE INDEX idx_categories_sort_order
ON categories(sort_order);
```

---

# 17. TABLE 13 — products

```sql
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
```

### Indexes

```sql
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
```

---

# 18. TABLE 14 — product_variants

```sql
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
```

### Indexes

```sql
CREATE INDEX idx_variants_product_id
ON product_variants(product_id);

CREATE INDEX idx_variants_barcode
ON product_variants(barcode)
WHERE barcode IS NOT NULL;
```

---

# 19. TABLE 15 — product_images

```sql
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
```

---

# 20. TABLE 16 — product_categories

Many-to-many product/category relation.

```sql
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
```

---

# 21. TABLE 17 — product_attributes

```sql
CREATE TABLE product_attributes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(100) NOT NULL UNIQUE,

    display_name VARCHAR(150) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Examples:

```text
Color
Size
RAM
Storage
Material
```

---

# 22. TABLE 18 — attribute_values

```sql
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
```

---

# 23. TABLE 19 — variant_attribute_values

```sql
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
```

---

# 24. TABLE 20 — inventory_locations

```sql
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
```

---

# 25. TABLE 21 — inventory

Current stock state.

```sql
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
```

Available stock:

```text
available = quantity - reserved_quantity
```

---

# 26. TABLE 22 — inventory_movements

Immutable stock ledger.

```sql
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
```

---

# 27. TABLE 23 — inventory_reservations

```sql
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
```

`order_id` will reference `orders` after the orders table exists.

---

# 28. TABLE 24 — carts

```sql
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
```

---

# 29. TABLE 25 — cart_items

```sql
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
```

---

# 30. TABLE 26 — wishlists

```sql
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
```

---

# 31. TABLE 27 — wishlist_items

```sql
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
```

---

# 32. TABLE 28 — orders

```sql
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
```

### Indexes

```sql
CREATE INDEX idx_orders_user_id
ON orders(user_id);

CREATE INDEX idx_orders_status
ON orders(status);

CREATE INDEX idx_orders_created_at
ON orders(created_at DESC);
```

---

# 33. TABLE 29 — order_items

Important: store snapshots of product information.

```sql
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
```

---

# 34. TABLE 30 — order_addresses

Address snapshots.

Do NOT depend on the user's current address for historical orders.

```sql
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
```

Possible:

```text
shipping
billing
```

---

# 35. TABLE 31 — order_status_history

```sql
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
```

---

# 36. TABLE 32 — payments

```sql
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
```

---

# 37. TABLE 33 — payment_transactions

```sql
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
```

---

# 38. TABLE 34 — refunds

```sql
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
```

---

# 39. TABLE 35 — shipping_methods

```sql
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
```

---

# 40. TABLE 36 — shipments

```sql
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
```

---

# 41. TABLE 37 — shipment_items

```sql
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
```

---

# 42. TABLE 38 — tracking_events

```sql
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
```

---

# 43. TABLE 39 — reviews

```sql
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
```

Prevent duplicate reviews for the same order item:

```sql
CREATE UNIQUE INDEX idx_reviews_unique_order_item
ON reviews(order_item_id);
```

---

# 44. TABLE 40 — review_images

```sql
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
```

---

# 45. TABLE 41 — review_votes

```sql
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
```

---

# 46. TABLE 42 — product_questions

```sql
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
```

---

# 47. TABLE 43 — product_question_answers

```sql
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
```

---

# 48. TABLE 44 — coupons

```sql
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
```

---

# 49. TABLE 45 — coupon_products

```sql
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
```

---

# 50. TABLE 46 — coupon_categories

```sql
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
```

---

# 51. TABLE 47 — returns

```sql
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
```

---

# 52. TABLE 48 — return_items

```sql
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
```

---

# 53. TABLE 49 — notifications

The full V2 implementation can include this table.

```sql
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
```

### Index

```sql
CREATE INDEX idx_notifications_user_created
ON notifications(user_id, created_at DESC);
```

---

# 54. TABLE 50 — notification_preferences

```sql
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
```

---

# 55. TABLE 51 — audit_logs

For production administration.

```sql
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
```

### Indexes

```sql
CREATE INDEX idx_audit_logs_user
ON audit_logs(user_id);

CREATE INDEX idx_audit_logs_entity
ON audit_logs(entity_type, entity_id);

CREATE INDEX idx_audit_logs_created_at
ON audit_logs(created_at DESC);
```

---

# 56. Correct Table Count

The complete schema above contains:

```text
1   users
2   user_profiles
3   user_sessions
4   user_addresses
5   roles
6   permissions
7   role_permissions
8   user_roles

9   vendors
10  vendor_documents

11  brands
12  categories
13  products
14  product_variants
15  product_images
16  product_categories
17  product_attributes
18  attribute_values
19  variant_attribute_values

20  inventory_locations
21  inventory
22  inventory_movements
23  inventory_reservations

24  carts
25  cart_items
26  wishlists
27  wishlist_items

28  orders
29  order_items
30  order_addresses
31  order_status_history

32  payments
33  payment_transactions
34  refunds

35  shipping_methods
36  shipments
37  shipment_items
38  tracking_events

39  reviews
40  review_images
41  review_votes
42  product_questions
43  product_question_answers

44  coupons
45  coupon_products
46  coupon_categories

47  returns
48  return_items

49  notifications
50  notification_preferences
51  audit_logs
```

**Total: 51 tables.**

---

# 57. Important Relationships

## User

```text
users 1 ─── 1 user_profiles

users 1 ─── N user_sessions

users 1 ─── N user_addresses

users N ─── N roles
       through user_roles
```

## Vendor

```text
users 1 ─── 1 vendors

vendors 1 ─── N vendor_documents

vendors 1 ─── N products
```

## Catalog

```text
products N ─── N categories
          through product_categories

products 1 ─── N product_variants

products 1 ─── N product_images

product_variants N ─── N attribute_values
```

## Inventory

```text
product_variants 1 ─── N inventory

inventory_locations 1 ─── N inventory

product_variants 1 ─── N inventory_movements
```

## Shopping

```text
users 1 ─── N carts

carts 1 ─── N cart_items

users 1 ─── 1 wishlists

wishlists 1 ─── N wishlist_items
```

## Orders

```text
users 1 ─── N orders

orders 1 ─── N order_items

orders 1 ─── N order_addresses

orders 1 ─── N order_status_history
```

## Payments

```text
orders 1 ─── N payments

payments 1 ─── N payment_transactions

payments 1 ─── N refunds
```

## Shipping

```text
orders 1 ─── N shipments

shipments 1 ─── N shipment_items

shipments 1 ─── N tracking_events
```

## Reviews

```text
products 1 ─── N reviews

users 1 ─── N reviews

reviews 1 ─── N review_images

reviews N ─── N users
through review_votes
```

## Promotions

```text
coupons N ─── N products

coupons N ─── N categories
```

## Returns

```text
orders 1 ─── N returns

returns 1 ─── N return_items
```

---

# 58. Critical Business Rules

## Product visibility

A product is publicly visible only when:

```text
product.status = 'active'
AND
product.approval_status = 'approved'
AND
vendor.status = 'approved'
```

---

## Vendor access

A vendor may access a product only when:

```text
product.vendor_id = authenticated_vendor_id
```

The same rule applies to:

```text
orders
inventory
reviews
questions
analytics
```

---

## Verified reviews

A user can review a product only if:

```text
order.user_id = authenticated_user_id
AND
order.status = delivered
AND
order_item.product_id = requested_product_id
```

Only one review per order item:

```text
UNIQUE(order_item_id)
```

---

## Inventory

Never allow:

```text
reserved_quantity > quantity
```

Available stock:

```text
available_quantity =
quantity - reserved_quantity
```

---

## Checkout

The backend must recalculate:

```text
subtotal
discount
tax
shipping
grand_total
```

Never trust frontend totals.

---

## Payments

An order is considered paid only after:

```text
verified payment provider response
```

or:

```text
verified payment webhook
```

Never mark an order as paid from a frontend request.

---

## Refunds

Total successful refunds must never exceed:

```text
total successful payment amount
```

---

# 59. Vendor Revenue Model

For marketplace accounting, don't derive vendor revenue from the current product price.

Use `order_items`.

Example:

```text
Product selling price = 1,000 BDT
Cost price = 700 BDT
Platform commission = 100 BDT
```

Then:

```text
Gross Sale          = 1,000
Cost of Goods       =   700
Platform Commission =   100
Vendor Net Revenue  =   900
Vendor Profit       =   200
Platform Revenue    =   100
```

For a real implementation, commission rules should eventually be configurable rather than hardcoded.

---

# 60. Recommended Additional Financial Table

For a serious marketplace, I recommend adding one more table beyond the 51-table core schema:

## vendor_ledger

```sql
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
```

Possible transaction types:

```text
sale
commission
refund
adjustment
payout
shipping_fee
```

This is valuable because a marketplace should have an **auditable financial ledger**, rather than calculating everything from dashboard queries.

That brings the recommended production schema to:

**52 tables.**

---

# 61. Recommended ERD Structure

```text
                                  ┌──────────────┐
                                  │    USERS     │
                                  └──────┬───────┘
                                         │
               ┌─────────────────────────┼────────────────────┐
               │                         │                    │
               ▼                         ▼                    ▼
        USER_PROFILES              USER_ADDRESSES       USER_SESSIONS
               │
               │
               ▼
            VENDORS
               │
        ┌──────┴───────────────┐
        │                      │
        ▼                      ▼
 VENDOR_DOCUMENTS          PRODUCTS
                              │
             ┌────────────────┼────────────────┐
             │                │                │
             ▼                ▼                ▼
       VARIANTS          CATEGORIES        IMAGES
             │
             ▼
         INVENTORY
             │
             ▼
   INVENTORY_MOVEMENTS


USERS
 │
 ├──────── CARTS ─────── CART_ITEMS ───── VARIANTS
 │
 ├──────── WISHLISTS ─── WISHLIST_ITEMS ─ PRODUCTS
 │
 └──────── ORDERS
                │
       ┌────────┼───────────┐
       │        │           │
       ▼        ▼           ▼
 ORDER_ITEMS  PAYMENTS    SHIPMENTS
       │        │           │
       │        ├── REFUNDS │
       │        └── TXNS    ├── SHIPMENT_ITEMS
       │                    └── TRACKING
       │
       ▼
    REVIEWS
       │
       ├── REVIEW_IMAGES
       └── REVIEW_VOTES


PRODUCTS
   │
   ├── QUESTIONS ─── QUESTION_ANSWERS
   │
   └── COUPONS
          ├── COUPON_PRODUCTS
          └── COUPON_CATEGORIES


ORDERS
   │
   └── RETURNS ─── RETURN_ITEMS


USERS
   │
   └── NOTIFICATIONS
          │
          └── NOTIFICATION_PREFERENCES


VENDORS
   │
   └── VENDOR_LEDGER


USERS
   │
   └── AUDIT_LOGS
```

---

# 62. Migration Order

The migration system should create tables in dependency order.

```text
001_extensions.sql

002_enums.sql

003_users.sql
004_user_profiles.sql
005_user_sessions.sql
006_user_addresses.sql

007_roles.sql
008_permissions.sql
009_role_permissions.sql
010_user_roles.sql

011_vendors.sql
012_vendor_documents.sql

013_brands.sql
014_categories.sql
015_products.sql
016_product_variants.sql
017_product_images.sql
018_product_categories.sql
019_product_attributes.sql
020_attribute_values.sql
021_variant_attribute_values.sql

022_inventory_locations.sql
023_inventory.sql
024_inventory_movements.sql

025_carts.sql
026_cart_items.sql
027_wishlists.sql
028_wishlist_items.sql

029_orders.sql
030_order_items.sql
031_order_addresses.sql
032_order_status_history.sql

033_inventory_reservations.sql

034_payments.sql
035_payment_transactions.sql
036_refunds.sql

037_shipping_methods.sql
038_shipments.sql
039_shipment_items.sql
040_tracking_events.sql

041_reviews.sql
042_review_images.sql
043_review_votes.sql
044_product_questions.sql
045_product_question_answers.sql

046_coupons.sql
047_coupon_products.sql
048_coupon_categories.sql

049_returns.sql
050_return_items.sql

051_notifications.sql
052_notification_preferences.sql

053_audit_logs.sql

054_vendor_ledger.sql

055_indexes.sql
056_seed_roles.sql
057_seed_permissions.sql
```

---

# 63. Important Implementation Note

There is one circular dependency:

```text
orders
   ↓
inventory_reservations
   ↓
order_id
```

Create `inventory_reservations` without the `order_id` foreign key initially, then add the constraint after `orders` exists:

```sql
ALTER TABLE inventory_reservations
ADD CONSTRAINT fk_reservations_order
FOREIGN KEY (order_id)
REFERENCES orders(id)
ON DELETE SET NULL;
```

This keeps migrations clean.

---

# 64. Database Design Principles

The AI coding agent MUST follow these principles:

1. PostgreSQL is the source of truth.
2. Never use floating point for money.
3. Never trust client-side prices.
4. Never trust client-side roles.
5. Enforce foreign keys.
6. Enforce critical constraints at database level.
7. Use transactions for checkout/payment/inventory.
8. Maintain inventory movement history.
9. Maintain order status history.
10. Maintain payment transaction history.
11. Maintain vendor financial ledger.
12. Store historical order snapshots.
13. Do not store product images in PostgreSQL.
14. Do not delete financial records casually.
15. Prefer soft state changes over destructive deletion for business-critical records.
16. Use indexes based on real query patterns.
17. Keep vendor data isolated.
18. Keep authentication separate from business profiles.
19. Make payment operations idempotent.
20. Make checkout idempotent.
21. Store all timestamps in UTC.
22. Use migrations for every schema change.
23. Never manually modify production schema.
24. Keep audit logs for privileged operations.
25. Do not expose internal database IDs unnecessarily in URLs if opaque public identifiers are introduced later.

---

# 65. Production Database Target

Final recommended architecture:

```text
                    ┌───────────────────┐
                    │      Next.js      │
                    └─────────┬─────────┘
                              │
                              ▼
                    ┌───────────────────┐
                    │       Go API      │
                    │                   │
                    │ JWT + RBAC        │
                    │ REST              │
                    │ Business Logic    │
                    └─────────┬─────────┘
                              │
              ┌───────────────┼────────────────┐
              │               │                │
              ▼               ▼                ▼
        ┌───────────┐   ┌───────────┐   ┌─────────────┐
        │ PostgreSQL│   │   Redis   │   │   Storage   │
        │   Neon    │   │           │   │ S3 / R2    │
        └───────────┘   └───────────┘   └─────────────┘
              │
              ▼
       52-table relational
       marketplace database
```

The database should remain a **single PostgreSQL database initially**. Do not split it into microservice databases until the traffic, team size, or operational requirements genuinely justify that complexity.

---

# 66. Store Applications, Memberships & Application Fee (migration 013)

Added by `migrations/013_store_applications.sql`.

## Enums

- `application_status` — `pending`, `approved`, `rejected`
- `store_status` — `pending_activation`, `active`, `suspended`, `closed`
- `store_member_role` — `owner`, `manager`
- `membership_status` — `active`, `invited`, `suspended`
- `business_type` — `individual`, `registered`
- Application-fee payments reuse the existing `payment_status` enum.

## Tables

- **`store_applications`** — the Facebook-Page-style application (basic info,
  contact, BD-structured address, business/verification fields) plus review
  metadata (`status`, `rejection_reason`, `admin_review_notes`, `reviewed_by`,
  `reviewed_at`). `application_number` and `store_slug` (active-only partial
  unique) are unique. Review data is kept here, separate from the live store.
- **`stores`** — the operational storefront, linked 1:1 to a `vendors` row
  (`stores.vendor_id` UNIQUE) and to its originating application
  (`stores.application_id` UNIQUE). `status` moves `pending_activation → active`.
- **`store_memberships`** — who may operate a store (`owner`/`manager`), unique
  per `(store_id, user_id)`. Store-scoped authorization layer; vendor
  operations resolve the caller's store through this table.
- **`application_payments`** — application-fee ledger, separate from the
  order-scoped `payments`. `payment_reference` unique; `provider_transaction_id`
  unique when present (idempotent webhook verification). Money is
  `NUMERIC(19,4)` + `currency CHAR(3)`.

## Changes to existing tables

- `vendors` gains `application_id`, `banner_url`, `category_id`; its
  `user_id` UNIQUE constraint is **dropped** (a user may own several stores),
  replaced by a plain index. `business_slug` remains globally unique.
- The demo vendor (seeded in 010) is backfilled with an active store and an
  owner membership.

## Rules

- Approval does not grant dashboard access. A store activates (and the `vendor`
  role is granted) only after the application is approved and, when the
  application fee is enabled, a payment is verified server-side.
- Payment status is kept separate from application status and store status.
- Activation is idempotent (unique `stores.application_id` + guarded state
  transitions). Client-supplied amounts/statuses are never trusted.

---

# 67. Vendor Operations (migration 018)

Added by `migrations/018_vendor_ops.sql`. Extends the existing vendor surface —
no new tables.

## Schema changes

- `products` gains `reviewed_at` / `reviewed_by`, set by every admin decision
  (`POST /admin/products/{id}/{approve|reject|changes}`), so a rejection or
  change request carries a review timestamp and reviewer. `approved_by` /
  `approved_at` remain approval-only.
- `inventory_movements` gains `quantity_before` / `quantity_after`, so the stock
  history shows the level before and after each change. Populated by the vendor
  stock adjustment, initial product stock and order-sale writers; older rows
  stay NULL.
- Index `idx_inventory_movements_variant_created (variant_id, created_at DESC)`
  backs the history view.

## Endpoints (all under `/api/v1/vendor`, store-scoped)

- `GET /inventory/movements` — stock history (paged; `variant_id`, `type`,
  `search` filters).
- `PUT /inventory/{variantId}/threshold` — set a variant's low-stock threshold.
- `POST /orders/{id}/status` — forward fulfilment transitions (`confirmed`,
  `processing`, `packed`, `shipped`, `delivered`); non-forward transitions are
  rejected (`INVALID_TRANSITION`) and terminal orders are locked.
- `GET /orders/export` and `GET /inventory/export` — CSV (default) or XLSX
  (`?format=`), vendor-scoped, with CSV formula-injection neutralisation and a
  10,000-row cap.

## Side effects

- **Notifications** — `internal/notify` writes in-app `notifications` rows.
  Stock alerts fire only on a worsening crossing (ok → low → out), never on an
  unchanged level. Order transitions notify the buyer (`order_confirmed`,
  `order_shipped`, `order_delivered`); a failed delivery attempt notifies too.
  In-app only: no email/SMS/push delivery worker exists yet.
- **Audit** — order status/shipment changes and inventory threshold changes
  write `audit_logs` rows.
- **Courier boundary** — `pkg/courier` defines a provider interface with a
  manual default (no fabricated tracking, no carrier API).
