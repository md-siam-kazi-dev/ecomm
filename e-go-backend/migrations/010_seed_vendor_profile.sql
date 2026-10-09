-- 010_seed_vendor_profile.sql
-- Give the demo vendor account a business profile so vendor endpoints can
-- resolve a vendor from the authenticated user. Idempotent.

INSERT INTO vendors (user_id, business_name, business_slug, business_email, status)
SELECT u.id, 'Vendor Demo Co', 'vendor-demo-co', u.email, 'approved'
FROM users u
WHERE u.email = 'vendor@vendor.com'
ON CONFLICT (user_id) DO NOTHING;
