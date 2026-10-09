-- 008_auth_tokens.sql
-- One-time tokens for email verification and password reset. Only the SHA-256
-- hash of each token is stored; the raw token is only ever sent to the user.

CREATE TYPE auth_token_purpose AS ENUM ('email_verification', 'password_reset');

CREATE TABLE auth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL,

    purpose auth_token_purpose NOT NULL,

    token_hash CHAR(64) NOT NULL UNIQUE,

    expires_at TIMESTAMPTZ NOT NULL,

    used_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_auth_tokens_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_auth_tokens_user_id ON auth_tokens(user_id);

CREATE INDEX idx_auth_tokens_expires_at ON auth_tokens(expires_at);
