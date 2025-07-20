ALTER TABLE events
  ADD COLUMN uuid uuid NOT NULL DEFAULT gen_random_uuid (),
  ADD CONSTRAINT events_uuid_unique UNIQUE (uuid);
