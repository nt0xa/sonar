ALTER TABLE payloads
  ALTER COLUMN store_events DROP DEFAULT,
  ALTER COLUMN store_events TYPE BOOL USING store_events > 0,
  ALTER COLUMN store_events SET DEFAULT false;

