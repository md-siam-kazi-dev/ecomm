-- 014_payments_verified_at.sql
-- The order/payment module records when a provider-verified payment was
-- confirmed. The payments table predates that field, so add it.

ALTER TABLE payments ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ;
