-- name: EventsCreate :one
INSERT INTO events (uuid, payload_id, protocol, r, w, rw, meta, remote_addr,
  received_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
RETURNING *;

-- name: EventsGetByID :one
SELECT * FROM events WHERE id = $1;

-- name: EventsListByPayloadID :many
SELECT 
  sqlc.embed(events),
  ROW_NUMBER() OVER(PARTITION BY payload_id ORDER BY id ASC) AS index
FROM events WHERE payload_id = $1
ORDER BY id DESC
LIMIT $2
OFFSET $3;

-- name: EventsGetByPayloadAndIndex :one
WITH numbered AS (
  SELECT
    id,
    ROW_NUMBER() OVER (PARTITION BY payload_id ORDER BY id ASC) AS index
  FROM events
)
SELECT sqlc.embed(e), n.index
FROM events e
JOIN numbered n ON n.id = e.id
WHERE e.payload_id = @payload_id::bigint AND n.index = @index::bigint;
