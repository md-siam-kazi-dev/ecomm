-- 016_review_images_index.sql
--
-- Lookup index for review photos. `review_images` already exists (migration
-- 003) but has no index on its foreign key; reviews are always read with their
-- images grouped by review_id.
--
-- Applied per-statement (no transaction wrapper), matching migrations 007-015.

CREATE INDEX IF NOT EXISTS idx_review_images_review
    ON review_images (review_id, sort_order);
