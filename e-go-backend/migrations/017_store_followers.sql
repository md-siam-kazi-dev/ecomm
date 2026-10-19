-- 017_store_followers.sql
--
-- Customers can follow a store; the store page shows the follower count.
-- One row per (store, user).
--
-- Applied per-statement (no transaction wrapper), matching migrations 007-016.

CREATE TABLE IF NOT EXISTS store_followers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_store_followers UNIQUE (store_id, user_id),
    CONSTRAINT fk_store_followers_store FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE,
    CONSTRAINT fk_store_followers_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_store_followers_store ON store_followers (store_id);
CREATE INDEX IF NOT EXISTS idx_store_followers_user ON store_followers (user_id);
