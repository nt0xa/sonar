BEGIN;

ALTER TABLE users ALTER COLUMN id TYPE int;
ALTER TABLE users ALTER COLUMN created_by TYPE int;
ALTER TABLE user_params ALTER COLUMN id TYPE int;
ALTER TABLE user_params ALTER COLUMN user_id TYPE int;
ALTER TABLE payloads ALTER COLUMN id TYPE int;
ALTER TABLE payloads ALTER COLUMN user_id TYPE int;
ALTER TABLE http_routes ALTER COLUMN id TYPE int;
ALTER TABLE http_routes ALTER COLUMN payload_id TYPE int;
ALTER TABLE dns_records ALTER COLUMN id TYPE int;
ALTER TABLE dns_records ALTER COLUMN payload_id TYPE int;
ALTER TABLE events ALTER COLUMN id TYPE int;
ALTER TABLE events ALTER COLUMN payload_id TYPE int;

COMMIT;
