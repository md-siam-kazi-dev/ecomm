-- 022_security_events.sql
-- API security monitoring: a persisted trail of the *interesting* API events
-- (authentication/authorisation failures, rate-limit rejections, server errors
-- and anomaly signals) plus the indexes the admin Security Monitoring screens
-- filter on. High-volume per-request metrics are aggregated in memory and are
-- deliberately NOT written here on the request path.
--
-- Kept separate from audit_logs on purpose: that table is the long-lived
-- *business* audit trail (who changed what), while this one is high-volume
-- operational telemetry with its own short retention policy.
--
-- Applied per-statement (no transaction wrapper), matching migrations 007-021.
-- Every statement is idempotent so re-running against an already-migrated
-- database is a no-op.

CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Stable identifier echoed to SSE clients and used for replay/dedupe.
    event_id TEXT NOT NULL,
    request_id VARCHAR(100),
    event_type VARCHAR(64) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    user_id UUID,
    user_role VARCHAR(20),
    ip_address INET,
    method VARCHAR(10),
    -- Route PATTERN (e.g. /api/v1/products/:id), never the raw URL with ids.
    endpoint TEXT,
    status_code SMALLINT,
    duration_ms INTEGER,
    user_agent TEXT,
    resource_type VARCHAR(64),
    resource_id UUID,
    -- Sanitized JSON only: bodies, tokens, cookies and credentials are never
    -- collected, and audit.Sanitize redacts anything sensitive that slips in.
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_security_events_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_security_events_severity
        CHECK (severity IN ('info', 'low', 'medium', 'high', 'critical')),
    CONSTRAINT uq_security_events_event_id UNIQUE (event_id)
);

-- Newest-first pagination with a stable tie-breaker.
CREATE INDEX IF NOT EXISTS idx_security_events_created_id
    ON security_events (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_security_events_type
    ON security_events (event_type);

CREATE INDEX IF NOT EXISTS idx_security_events_severity
    ON security_events (severity);

-- Drives "what did this address do?" investigations and the IP filter.
CREATE INDEX IF NOT EXISTS idx_security_events_ip_created
    ON security_events (ip_address, created_at DESC);

-- Sparse: most rows have no authenticated user.
CREATE INDEX IF NOT EXISTS idx_security_events_user
    ON security_events (user_id, created_at DESC)
    WHERE user_id IS NOT NULL;

-- Sparse: correlates a security event with the audit trail / request logs.
CREATE INDEX IF NOT EXISTS idx_security_events_request_id
    ON security_events (request_id)
    WHERE request_id IS NOT NULL;
