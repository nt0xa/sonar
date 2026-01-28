-- name: HTTPRoutesCreate :one
INSERT INTO http_routes (payload_id, method, path, code, headers, body,
  is_dynamic, created_at, index)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(),
  (SELECT COALESCE(MAX(index), 0) FROM http_routes hr WHERE hr.payload_id = $1) + 1)
RETURNING *;

-- name: HTTPRoutesUpdate :one
UPDATE http_routes SET
  payload_id = $2,
  method = $3,
  path = $4,
  code = $5,
  headers = $6,
  body = $7,
  is_dynamic = $8
WHERE id = $1
RETURNING *;

-- name: HTTPRoutesGetByID :one
SELECT * FROM http_routes WHERE id = $1;

-- name: HTTPRoutesGetByPayloadID :many
SELECT * FROM http_routes WHERE payload_id = $1;

-- name: HTTPRoutesGetByPayloadMethodAndPath :one
SELECT * FROM http_routes
WHERE payload_id = $1 AND method = $2 AND path = $3;

-- name: HTTPRoutesGetByPayloadIDAndIndex :one
SELECT * FROM http_routes WHERE payload_id = $1 AND index = $2;

-- name: HTTPRoutesDelete :exec
DELETE FROM http_routes WHERE id = $1;

-- name: HTTPRoutesDeleteAllByPayloadID :many
DELETE FROM http_routes WHERE payload_id = $1 RETURNING *;

-- name: HTTPRoutesDeleteAllByPayloadIDAndPath :many
DELETE FROM http_routes WHERE payload_id = $1 AND path = $2 RETURNING *;
