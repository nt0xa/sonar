-- name: DNSRecordsCreate :one
INSERT INTO dns_records (payload_id, name, type, ttl, values, strategy,
  last_answer, last_accessed_at, created_at, index)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(),
  (SELECT COALESCE(MAX(index), 0) FROM dns_records dr WHERE dr.payload_id = $1) + 1)
RETURNING *;

-- name: DNSRecordsUpdate :one
UPDATE dns_records SET
  payload_id = $2,
  name = $3,
  type = $4,
  ttl = $5,
  values = $6,
  strategy = $7,
  last_answer = $8,
  last_accessed_at = $9
WHERE id = $1
RETURNING *;

-- name: DNSRecordsGetByID :one
SELECT * FROM dns_records WHERE id = $1;

-- name: DNSRecordsGetByPayloadNameAndType :one
SELECT * FROM dns_records
WHERE payload_id = $1 AND name = $2 AND type = $3;

-- name: DNSRecordsGetByPayloadID :many
SELECT * FROM dns_records WHERE payload_id = $1 ORDER BY id ASC;

-- name: DNSRecordsGetCountByPayloadID :one
SELECT COUNT(*) FROM dns_records WHERE payload_id = $1;

-- name: DNSRecordsGetByPayloadIDAndIndex :one
SELECT * FROM dns_records WHERE payload_id = $1 AND index = $2;

-- name: DNSRecordsDelete :exec
DELETE FROM dns_records WHERE id = $1;

-- name: DNSRecordsDeleteAllByPayloadID :many
DELETE FROM dns_records WHERE payload_id = $1 RETURNING *;

-- name: DNSRecordsDeleteAllByPayloadIDAndName :many
DELETE FROM dns_records WHERE payload_id = $1 AND name = $2 RETURNING *;
