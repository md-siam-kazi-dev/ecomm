-- 020_audit_logs.sql
-- Enrich the existing audit_logs table (created in 003) for the admin Audit
-- Logs feature: actor role, outcome, failure reason, request correlation id
-- and structured non-sensitive metadata, plus indexes for the admin filters.
--
-- Applied per-statement (no transaction wrapper), matching migrations 007-019.
-- Every statement is idempotent so re-running against an already-migrated
-- database is a no-op.

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS actor_role VARCHAR(20);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS result VARCHAR(20);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS reason TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS request_id VARCHAR(100);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS metadata JSONB;

-- Stable ordering for pagination: newest first, id as the tie-breaker so two
-- events in the same microsecond never reorder between pages.
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_id
    ON audit_logs (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_action
    ON audit_logs (action);

CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_role
    ON audit_logs (actor_role);

CREATE INDEX IF NOT EXISTS idx_audit_logs_result
    ON audit_logs (result);

-- Only populated for HTTP-originated events; keep the index sparse.
CREATE INDEX IF NOT EXISTS idx_audit_logs_request_id
    ON audit_logs (request_id)
    WHERE request_id IS NOT NULL;
