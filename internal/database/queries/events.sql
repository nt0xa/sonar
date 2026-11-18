-- name: EventsCreate :one
WITH seq AS (
  UPDATE payloads SET last_event_index = last_event_index + 1
  WHERE id = $2
  RETURNING last_event_index
)
INSERT INTO events (uuid, payload_id, index, protocol, data, meta, remote_addr,
  received_at, created_at)
VALUES ($1, $2, (SELECT last_event_index FROM seq), $3, $4, $5, $6, $7, now())
RETURNING *;

-- name: EventsGetByID :one
SELECT * FROM events WHERE id = $1;

-- name: EventsListByPayloadID :many
SELECT * FROM events
WHERE payload_id = @payload_id AND index < @before
ORDER BY index DESC
LIMIT @count
OFFSET @page_offset;

-- name: EventsListByPayloadIDAfter :many
SELECT * FROM events
WHERE payload_id = @payload_id AND index > @after
ORDER BY index ASC
LIMIT @count;

-- name: EventsGetByPayloadAndIndex :one
SELECT * FROM events WHERE payload_id = $1 AND index = $2;
