-- 011_address_divisions.sql
-- Structured Bangladesh address fields for user_addresses. The generic
-- address_line_1 / city columns become optional (the structured fields replace
-- them) while full_name and phone stay required.

ALTER TABLE user_addresses ADD COLUMN IF NOT EXISTS division VARCHAR(100);
ALTER TABLE user_addresses ADD COLUMN IF NOT EXISTS district VARCHAR(100);
ALTER TABLE user_addresses ADD COLUMN IF NOT EXISTS upazila VARCHAR(100);
ALTER TABLE user_addresses ADD COLUMN IF NOT EXISTS village_town VARCHAR(150);
ALTER TABLE user_addresses ADD COLUMN IF NOT EXISTS clarification TEXT;

ALTER TABLE user_addresses ALTER COLUMN address_line_1 DROP NOT NULL;
ALTER TABLE user_addresses ALTER COLUMN city DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_user_addresses_user_id ON user_addresses(user_id);
