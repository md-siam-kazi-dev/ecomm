-- 022_security_monitoring.sql
-- API security monitoring: request/security events, aggregated alerts,
-- alert↔event correlation and IP restrictions.
--
-- Applied per-statement (no transaction wrapper), matching migrations 007-021.
-- Every statement is idempotent so re-running against an already-migrated
-- database is a no-op.

CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL UNIQUE,
    request_id VARCHAR(100),
    event_type VARCHAR(60) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    user_id UUID,
    user_role VARCHAR(20),
    ip_address INET,
    method VARCHAR(10),
    endpoint VARCHAR(255),
    status_code INTEGER,
    duration_ms INTEGER,
    user_agent TEXT,
    resource_type VARCHAR(40),
    resource_id VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_security_events_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_security_events_severity CHECK (severity IN ('info', 'low', 'medium', 'high', 'critical'))
);

-- Newest-first keyset pagination and range scans.
CREATE INDEX IF NOT EXISTS idx_security_events_created_id
    ON security_events (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_security_events_type_created
    ON security_events (event_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_security_events_severity_created
    ON security_events (severity, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_security_events_user
    ON security_events (user_id, created_at DESC)
    WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_security_events_ip
    ON security_events (ip_address, created_at DESC)
    WHERE ip_address IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_security_events_request_id
    ON security_events (request_id)
    WHERE request_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_security_events_status
    ON security_events (status_code, created_at DESC)
    WHERE status_code IS NOT NULL;

-- Aggregated alerts. dedupe_key identifies the pattern (rule + source) so
-- repeated detections update one active alert instead of creating a row per
-- request. The partial unique index only applies to active alerts, so a
-- resolved alert can be re-opened later as a new row if the pattern recurs.
CREATE TABLE IF NOT EXISTS security_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dedupe_key TEXT NOT NULL,
    event_type VARCHAR(60) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    source_ip INET,
    user_id UUID,
    endpoint VARCHAR(255),
    first_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    occurrence_count BIGINT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    notes TEXT,
    resolved_at TIMESTAMPTZ,
    resolved_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_security_alerts_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_security_alerts_resolved_by FOREIGN KEY (resolved_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_security_alerts_severity CHECK (severity IN ('info', 'low', 'medium', 'high', 'critical')),
    CONSTRAINT chk_security_alerts_status CHECK (status IN ('open', 'investigating', 'resolved', 'false_positive'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_security_alerts_active_dedupe
    ON security_alerts (dedupe_key)
    WHERE status IN ('open', 'investigating');
CREATE INDEX IF NOT EXISTS idx_security_alerts_status_last_seen
    ON security_alerts (status, last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_security_alerts_severity_last_seen
    ON security_alerts (severity, last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_security_alerts_last_seen
    ON security_alerts (last_seen DESC, id DESC);

-- Correlate the individual events that contributed to an aggregated alert.
CREATE TABLE IF NOT EXISTS security_alert_events (
    alert_id UUID NOT NULL,
    event_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (alert_id, event_id),
    CONSTRAINT fk_security_alert_events_alert FOREIGN KEY (alert_id) REFERENCES security_alerts(id) ON DELETE CASCADE,
    CONSTRAINT fk_security_alert_events_event FOREIGN KEY (event_id) REFERENCES security_events(event_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_security_alert_events_event
    ON security_alert_events (event_id);

-- IP restrictions. Rows are immutable history: unblocking sets revoked_at
-- instead of deleting, so the audit trail survives. An active row is one with
-- revoked_at IS NULL and (permanent OR expires_at > NOW()).
CREATE TABLE IF NOT EXISTS ip_restrictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address INET NOT NULL,
    reason TEXT NOT NULL,
    is_permanent BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    revoked_by UUID,
    metadata JSONB,

    CONSTRAINT fk_ip_restrictions_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_ip_restrictions_revoked_by FOREIGN KEY (revoked_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_ip_restrictions_expiry CHECK (is_permanent OR expires_at IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_ip_restrictions_ip_active
    ON ip_restrictions (ip_address)
    WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ip_restrictions_created
    ON ip_restrictions (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_ip_restrictions_active_expiry
    ON ip_restrictions (expires_at)
    WHERE revoked_at IS NULL AND is_permanent = FALSE;
