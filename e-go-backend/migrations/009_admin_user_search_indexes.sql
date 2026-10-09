-- 009_admin_user_search_indexes.sql
-- Indexes backing the admin user-management search/filter/sort.
-- Trigram (pg_trgm) GIN indexes make `ILIKE '%term%'` partial matching usable.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_users_email_trgm
    ON users USING gin (email gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_user_profiles_first_name_trgm
    ON user_profiles USING gin (first_name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_user_profiles_last_name_trgm
    ON user_profiles USING gin (last_name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_users_created_at
    ON users (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_users_last_login_at
    ON users (last_login_at DESC);

-- Role filtering joins roles → user_roles on role_id; the composite PK is
-- (user_id, role_id), so add a reverse index for the join.
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id
    ON user_roles (role_id);
