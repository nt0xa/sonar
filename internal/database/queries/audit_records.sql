-- name: AuditRecordsCreate :one
INSERT INTO audit_records (
  operation,
  actor,
  target,
  data
)
VALUES (
  @operation,
  @actor,
  @target,
  @data
)
RETURNING *;

-- name: AuditRecordsGetByID :one
SELECT * FROM audit_records WHERE id = @id;

-- name: AuditRecordsList :many
SELECT * FROM audit_records
WHERE
  (sqlc.narg(actor_id)::bigint IS NULL OR NULLIF(actor->>'id', '')::bigint = sqlc.narg(actor_id)::bigint)
  AND (@actor_name::text = '' OR actor->>'name' = @actor_name::text)
  AND (@resource_type::text = '' OR target->>'type' = @resource_type::text)
  AND (sqlc.narg(resource_id)::bigint IS NULL OR NULLIF(target->>'id', '')::bigint = sqlc.narg(resource_id)::bigint)
  AND (@resource_key::text = '' OR target->>'key' = @resource_key::text)
  AND (@action::text = '' OR operation::text = @action::text)
  AND (sqlc.narg(payload_id)::bigint IS NULL OR NULLIF(target->>'payload_id', '')::bigint = sqlc.narg(payload_id)::bigint)
  AND (@payload_name::text = '' OR target->>'payload_name' = @payload_name::text)
  AND (sqlc.narg(from_at)::timestamptz IS NULL OR created_at >= sqlc.narg(from_at)::timestamptz)
  AND (sqlc.narg(to_at)::timestamptz IS NULL OR created_at <= sqlc.narg(to_at)::timestamptz)
ORDER BY id DESC
LIMIT @page_limit
OFFSET @page_offset;
