CREATE TYPE audit_record_operation_type AS ENUM ('create', 'update', 'delete', 'clear');

CREATE TABLE audit_records (
  id bigint NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  operation audit_record_operation_type NOT NULL,
  actor JSONB NOT NULL DEFAULT '{}',
  target JSONB NOT NULL DEFAULT '{}',
  data JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX audit_records_created_at_idx ON audit_records (created_at DESC);
CREATE INDEX audit_records_operation_created_at_idx ON audit_records (operation, created_at DESC);
CREATE INDEX audit_records_target_gin_idx ON audit_records USING GIN (target jsonb_path_ops);
CREATE INDEX audit_records_target_type_created_at_idx ON audit_records ((target->>'type'), created_at DESC);
