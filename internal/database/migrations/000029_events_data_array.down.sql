BEGIN;

ALTER TABLE events
  ADD COLUMN r bytea,
  ADD COLUMN w bytea,
  ADD COLUMN rw bytea;

UPDATE events SET
  r  = COALESCE(data[1], ''),
  w  = COALESCE(data[2], ''),
  rw = COALESCE((SELECT string_agg(m, ''::bytea) FROM unnest(data) AS m), '');

ALTER TABLE events
  ALTER COLUMN r SET NOT NULL,
  ALTER COLUMN w SET NOT NULL,
  ALTER COLUMN rw SET NOT NULL;

ALTER TABLE events DROP COLUMN data;

COMMIT;
