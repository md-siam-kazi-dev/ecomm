-- 012_product_approval.sql
-- Vendor product submission + admin approval workflow.
--
-- New approval state `changes_requested`, review metadata on products, an
-- optional attributes payload, and the storefront categories seeded so vendors
-- have something to pick from.

ALTER TYPE approval_status ADD VALUE IF NOT EXISTS 'changes_requested';

ALTER TABLE products ADD COLUMN IF NOT EXISTS review_note TEXT;
ALTER TABLE products ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ;
ALTER TABLE products ADD COLUMN IF NOT EXISTS attributes JSONB;

INSERT INTO categories (name, slug) VALUES
    ('Electronics & Gadgets', 'electronics'),
    ('Home & Kitchen', 'home-kitchen'),
    ('Fashion & Accessories', 'fashion'),
    ('Beauty & Care', 'beauty'),
    ('Sports & Fitness', 'sports'),
    ('Toys & Games', 'toys'),
    ('Pet Supplies', 'pets'),
    ('Office & Stationery', 'office')
ON CONFLICT (slug) DO NOTHING;
