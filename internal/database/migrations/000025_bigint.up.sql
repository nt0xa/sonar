BEGIN;

ALTER TABLE users ALTER COLUMN id TYPE bigint;
ALTER TABLE users ALTER COLUMN created_by TYPE bigint;
ALTER TABLE user_params ALTER COLUMN id TYPE bigint;
ALTER TABLE user_params ALTER COLUMN user_id TYPE bigint;
ALTER TABLE payloads ALTER COLUMN id TYPE bigint;
ALTER TABLE payloads ALTER COLUMN user_id TYPE bigint;
ALTER TABLE http_routes ALTER COLUMN id TYPE bigint;
ALTER TABLE http_routes ALTER COLUMN payload_id TYPE bigint;
ALTER TABLE dns_records ALTER COLUMN id TYPE bigint;
ALTER TABLE dns_records ALTER COLUMN payload_id TYPE bigint;
ALTER TABLE events ALTER COLUMN id TYPE bigint;
ALTER TABLE events ALTER COLUMN payload_id TYPE bigint;

COMMIT;
