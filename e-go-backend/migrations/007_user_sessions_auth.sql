-- 007_user_sessions_auth.sql
-- Session metadata + lookup index required for refresh-token rotation.

ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_sessions_refresh_token_hash
ON user_sessions(refresh_token_hash);
