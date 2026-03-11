-- name: AuditRecordsCreate :one
INSERT INTO audit_records (
  action,
  resource_type,
  source,
  actor_id,
  actor_name,
  actor_metadata,
  resource
)
VALUES (
  @action,
  @resource_type,
  @source,
  @actor_id,
  @actor_name,
  @actor_metadata,
  @resource
)
RETURNING *;

-- name: AuditRecordsGetByID :one
SELECT * FROM audit_records WHERE id = @id;

-- name: AuditRecordsList :many
SELECT * FROM audit_records
WHERE
  (sqlc.narg(actor_id)::bigint IS NULL OR actor_id = sqlc.narg(actor_id)::bigint)
  AND (@actor_name::text = '' OR actor_name = @actor_name::text)
  AND (@resource_type::text = '' OR resource_type::text = @resource_type::text)
  AND (@action::text = '' OR action::text = @action::text)
  AND (sqlc.narg(from_at)::timestamptz IS NULL OR created_at >= sqlc.narg(from_at)::timestamptz)
  AND (sqlc.narg(to_at)::timestamptz IS NULL OR created_at <= sqlc.narg(to_at)::timestamptz)
ORDER BY id DESC
LIMIT @page_limit
OFFSET @page_offset;
