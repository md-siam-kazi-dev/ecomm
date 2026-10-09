-- 002_enums.sql
-- PostgreSQL enum types. Create all types before any table that uses them.

CREATE TYPE user_status AS ENUM ('pending', 'active', 'suspended', 'blocked');

CREATE TYPE user_role_type AS ENUM ('user', 'vendor', 'admin');

CREATE TYPE vendor_status AS ENUM ('pending', 'approved', 'rejected', 'suspended');

CREATE TYPE product_status AS ENUM ('draft', 'active', 'inactive', 'archived', 'suspended');

CREATE TYPE approval_status AS ENUM ('pending', 'approved', 'rejected');

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

CREATE TYPE cart_status AS ENUM ('active', 'converted', 'abandoned', 'expired');

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

CREATE TYPE payment_status AS ENUM (
    'pending',
    'processing',
    'paid',
    'failed',
    'cancelled',
    'partially_refunded',
    'refunded'
);

CREATE TYPE payment_transaction_type AS ENUM (
    'authorization',
    'capture',
    'payment',
    'refund',
    'partial_refund',
    'void'
);

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

CREATE TYPE review_status AS ENUM ('pending', 'published', 'rejected', 'hidden');

CREATE TYPE coupon_type AS ENUM ('percentage', 'fixed_amount', 'free_shipping');

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

CREATE TYPE notification_channel AS ENUM ('in_app', 'email', 'sms');
