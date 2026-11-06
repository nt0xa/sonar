-- Backfill true to match the always-store behaviour, then restore the old default.
ALTER TABLE payloads ADD COLUMN store_events BOOL NOT NULL DEFAULT true;
ALTER TABLE payloads ALTER COLUMN store_events SET DEFAULT false;
