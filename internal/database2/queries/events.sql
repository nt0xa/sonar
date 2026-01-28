-- name: EventsCreate :one
INSERT INTO events (uuid, payload_id, protocol, r, w, rw, meta, remote_addr,
  received_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
RETURNING *;

-- name: EventsGetByID :one
SELECT * FROM events WHERE id = $1;

-- name: EventsListByPayloadID :many
SELECT * FROM events WHERE payload_id = $1
ORDER BY id DESC LIMIT $2 OFFSET $3;

-- name: EventsGetByPayloadAndIndex :one
SELECT *
FROM (
  SELECT events.*, ROW_NUMBER() OVER(ORDER BY id ASC) AS index
  FROM events WHERE payload_id = sqlc.arg(payload_id)::bigint
) subq WHERE index = sqlc.arg(index)::bigint;


