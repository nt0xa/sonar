BEGIN;

ALTER TABLE users ALTER COLUMN created_at TYPE timestamptz USING created_at AT TIME ZONE 'UTC';
ALTER TABLE payloads ALTER COLUMN created_at TYPE timestamptz USING created_at AT TIME ZONE 'UTC';
ALTER TABLE dns_records ALTER COLUMN created_at TYPE timestamptz USING created_at AT TIME ZONE 'UTC';
ALTER TABLE dns_records ALTER COLUMN last_accessed_at TYPE timestamptz USING last_accessed_at AT TIME ZONE 'UTC';
ALTER TABLE http_routes ALTER COLUMN created_at TYPE timestamptz USING created_at AT TIME ZONE 'UTC';
ALTER TABLE events ALTER COLUMN created_at TYPE timestamptz USING created_at AT TIME ZONE 'UTC';
ALTER TABLE events ALTER COLUMN received_at TYPE timestamptz USING received_at AT TIME ZONE 'UTC';

COMMIT;
