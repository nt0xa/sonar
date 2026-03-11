CREATE TYPE audit_record_action_type AS ENUM ('create', 'update', 'delete');
CREATE TYPE audit_record_resource_type AS ENUM ('payload', 'user', 'dns_record', 'http_route');
CREATE TYPE audit_record_source_type AS ENUM ('api', 'telegram', 'lark', 'slack');

CREATE TABLE audit_records (
  id bigint NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  uuid UUID NOT NULL DEFAULT gen_random_uuid(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  action audit_record_action_type NOT NULL,
  resource_type audit_record_resource_type NOT NULL,
  source audit_record_source_type NOT NULL,
  actor_id bigint,
  actor_name text NOT NULL DEFAULT '',
  actor_metadata JSONB NOT NULL DEFAULT '{}',
  resource JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX audit_records_created_at_idx ON audit_records (created_at DESC);
CREATE UNIQUE INDEX audit_records_uuid_idx ON audit_records (uuid);
CREATE INDEX audit_records_action_created_at_idx ON audit_records (action, created_at DESC);
CREATE INDEX audit_records_resource_type_created_at_idx ON audit_records (resource_type, created_at DESC);
CREATE INDEX audit_records_actor_id_created_at_idx ON audit_records (actor_id, created_at DESC);
CREATE INDEX audit_records_actor_name_created_at_idx ON audit_records (actor_name, created_at DESC);
CREATE INDEX audit_records_resource_gin_idx ON audit_records USING GIN (resource jsonb_path_ops);
