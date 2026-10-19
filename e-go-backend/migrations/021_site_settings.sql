-- 021_site_settings.sql
-- Admin-editable storefront settings (currently the announcement bar shown
-- above the header). A single row holds the configuration; the fixed primary
-- key plus a CHECK constraint guarantees it can never grow beyond one row.
--
-- Applied per-statement (no transaction wrapper), matching migrations 007-020.
-- Every statement is idempotent so re-running against an already-migrated
-- database is a no-op.

CREATE TABLE IF NOT EXISTS site_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1,
    announcement_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    announcement_message TEXT NOT NULL DEFAULT 'Free shipping on orders over $75 · 30-day easy returns',
    updated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_site_settings_singleton CHECK (id = 1),
    CONSTRAINT fk_site_settings_updated_by FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Seed the single settings row so the storefront always has a value to read.
INSERT INTO site_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;
