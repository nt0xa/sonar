-- name: PayloadsCreate :one
INSERT INTO payloads (user_id, subdomain, name, notify_protocols, store_events, created_at)
VALUES ($1, $2, $3, $4, $5, now())
RETURNING *;

-- name: PayloadsUpdate :one
UPDATE payloads SET
  subdomain = $2,
  user_id = $3,
  name = $4,
  notify_protocols = $5,
  store_events = $6
WHERE id = $1
RETURNING *;

-- name: PayloadsGetByID :one
SELECT * FROM payloads WHERE id = $1;

-- name: PayloadsGetBySubdomain :one
SELECT * FROM payloads WHERE subdomain = $1;

-- name: PayloadsGetByUserAndName :one
SELECT * FROM payloads WHERE user_id = $1 AND name = $2;

-- name: PayloadsFindByUserID :many
SELECT * FROM payloads WHERE user_id = $1 ORDER BY created_at DESC;

-- name: PayloadsFindByUserAndName :many
SELECT * FROM payloads
WHERE user_id = $1 AND name ILIKE '%' || @name::text || '%'
ORDER BY id DESC LIMIT $3 OFFSET $4;

-- name: PayloadsDelete :one
DELETE FROM payloads WHERE id = $1 RETURNING *;

-- name: PayloadsDeleteByNamePart :many
DELETE FROM payloads WHERE user_id = $1 AND name ILIKE '%' || @name::text || '%' RETURNING *;

-- name: PayloadsGetAllSubdomains :many
SELECT subdomain FROM payloads;
