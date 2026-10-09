-- 013_store_applications.sql
-- Store application -> admin review -> application-fee -> store activation.
--
-- Adds the "store" entity plus store memberships (multi-store, store-scoped
-- vendor authorization) and the application-fee payment ledger. The existing
-- `vendors` table is retained as the catalog anchor (products/orders key off
-- vendor_id); each store is linked 1:1 to a vendor. Review data stays on the
-- application so the operational store profile remains clean.
--
-- Applied per-statement (no transaction wrapper), matching migrations 007-012.

CREATE TYPE application_status AS ENUM ('pending', 'approved', 'rejected');

CREATE TYPE store_status AS ENUM ('pending_activation', 'active', 'suspended', 'closed');

CREATE TYPE store_member_role AS ENUM ('owner', 'manager');

CREATE TYPE membership_status AS ENUM ('active', 'invited', 'suspended');

CREATE TYPE business_type AS ENUM ('individual', 'registered');

-- Store applications. Review + audit data, kept separate from the live store.
CREATE TABLE store_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_number VARCHAR(50) NOT NULL,
    applicant_user_id UUID NOT NULL,
    store_name VARCHAR(255) NOT NULL,
    store_slug VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category_id UUID,

    logo_url TEXT,
    banner_url TEXT,

    business_email VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(30) NOT NULL,
    support_phone VARCHAR(30),
    website_url TEXT,

    country VARCHAR(100) NOT NULL DEFAULT 'Bangladesh',
    division VARCHAR(100) NOT NULL,
    district VARCHAR(100) NOT NULL,
    full_address TEXT NOT NULL,
    postal_code VARCHAR(30),

    business_type business_type NOT NULL DEFAULT 'individual',
    registered_business_name VARCHAR(255),
    trade_license_number VARCHAR(100),

    status application_status NOT NULL DEFAULT 'pending',
    rejection_reason TEXT,
    admin_review_notes TEXT,
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,

    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_store_applications_number UNIQUE (application_number),
    CONSTRAINT fk_store_applications_applicant FOREIGN KEY (applicant_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT fk_store_applications_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
    CONSTRAINT fk_store_applications_reviewer FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_store_applications_applicant ON store_applications (applicant_user_id);
CREATE INDEX IF NOT EXISTS idx_store_applications_status ON store_applications (status);
CREATE INDEX IF NOT EXISTS idx_store_applications_submitted_at ON store_applications (submitted_at DESC);

-- One active application per store slug; rejected applications don't reserve it.
CREATE UNIQUE INDEX IF NOT EXISTS idx_store_applications_slug_active
    ON store_applications (store_slug)
    WHERE status IN ('pending', 'approved');

CREATE INDEX IF NOT EXISTS idx_store_applications_store_name_trgm
    ON store_applications USING gin (store_name gin_trgm_ops);

-- Operational storefront. Backed by one vendor row (catalog anchor).
CREATE TABLE stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id UUID NOT NULL,
    application_id UUID,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID,
    logo_url TEXT,
    banner_url TEXT,
    status store_status NOT NULL DEFAULT 'pending_activation',
    activated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_stores_vendor UNIQUE (vendor_id),
    CONSTRAINT uq_stores_application UNIQUE (application_id),
    CONSTRAINT uq_stores_slug UNIQUE (slug),
    CONSTRAINT fk_stores_vendor FOREIGN KEY (vendor_id) REFERENCES vendors(id) ON DELETE RESTRICT,
    CONSTRAINT fk_stores_application FOREIGN KEY (application_id) REFERENCES store_applications(id) ON DELETE SET NULL,
    CONSTRAINT fk_stores_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_stores_status ON stores (status);
CREATE INDEX IF NOT EXISTS idx_stores_slug ON stores (slug);

-- Who may operate a store (owner/manager). Store-scoped authorization.
CREATE TABLE store_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role store_member_role NOT NULL DEFAULT 'owner',
    status membership_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_store_membership UNIQUE (store_id, user_id),
    CONSTRAINT fk_store_memberships_store FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE,
    CONSTRAINT fk_store_memberships_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_store_memberships_user ON store_memberships (user_id);

-- Application-fee payment ledger (separate from the order-scoped `payments`).
CREATE TABLE application_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL,
    store_id UUID,
    user_id UUID NOT NULL,
    amount NUMERIC(19,4) NOT NULL,
    currency CHAR(3) NOT NULL,
    provider VARCHAR(100),
    payment_reference VARCHAR(100) NOT NULL,
    provider_transaction_id VARCHAR(255),
    status payment_status NOT NULL DEFAULT 'pending',
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_application_payments_reference UNIQUE (payment_reference),
    CONSTRAINT chk_application_payment_amount CHECK (amount >= 0),
    CONSTRAINT fk_application_payments_application FOREIGN KEY (application_id) REFERENCES store_applications(id) ON DELETE CASCADE,
    CONSTRAINT fk_application_payments_store FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL,
    CONSTRAINT fk_application_payments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_application_payments_application ON application_payments (application_id);
CREATE INDEX IF NOT EXISTS idx_application_payments_user ON application_payments (user_id);

-- Provider transaction ids are globally unique when present (webhook idempotency).
CREATE UNIQUE INDEX IF NOT EXISTS idx_application_payments_provider_txn
    ON application_payments (provider_transaction_id)
    WHERE provider_transaction_id IS NOT NULL;

-- The `vendors` table becomes the catalog anchor for a store: a user may now
-- own several stores, so the 1:1 user constraint is removed and replaced by a
-- plain index. business_slug stays globally unique.
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS application_id UUID;
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS banner_url TEXT;
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS category_id UUID;

ALTER TABLE vendors DROP CONSTRAINT IF EXISTS vendors_user_id_key;
CREATE INDEX IF NOT EXISTS idx_vendors_user_id ON vendors (user_id);

-- Backfill: give the existing demo vendor (seeded in 010) a store + owner
-- membership so the vendor dashboard keeps working under the new model.
INSERT INTO stores (vendor_id, name, slug, description, logo_url, status, activated_at)
SELECT v.id, v.business_name, v.business_slug, v.description, v.logo_url, 'active', NOW()
FROM vendors v
ON CONFLICT (vendor_id) DO NOTHING;

INSERT INTO store_memberships (store_id, user_id, role, status)
SELECT s.id, v.user_id, 'owner', 'active'
FROM stores s
JOIN vendors v ON v.id = s.vendor_id
ON CONFLICT (store_id, user_id) DO NOTHING;
