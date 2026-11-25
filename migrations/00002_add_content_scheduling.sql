-- +goose Up
-- Add scheduled_at to content_translations
ALTER TABLE content_translations ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ;

-- Drop old status check constraint and add updated constraint including 'scheduled'
ALTER TABLE content_translations DROP CONSTRAINT IF EXISTS content_translations_status_chk;
ALTER TABLE content_translations ADD CONSTRAINT content_translations_status_chk
    CHECK (status IN ('draft', 'scheduled', 'published', 'archived'));

-- Index for querying mature scheduled content
CREATE INDEX IF NOT EXISTS content_translations_scheduled_idx
    ON content_translations (scheduled_at)
    WHERE status = 'scheduled';

-- +goose Down
DROP INDEX IF EXISTS content_translations_scheduled_idx;
ALTER TABLE content_translations DROP CONSTRAINT IF EXISTS content_translations_status_chk;
ALTER TABLE content_translations ADD CONSTRAINT content_translations_status_chk
    CHECK (status IN ('draft', 'published', 'archived'));
ALTER TABLE content_translations DROP COLUMN IF EXISTS scheduled_at;
