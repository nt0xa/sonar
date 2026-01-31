-- name: UsersCreate :one
INSERT INTO users (name, is_admin, created_by, api_token, telegram_id, lark_id, slack_id, created_at)
VALUES ($1, $2, $3, COALESCE(NULLIF(@api_token, ''), encode(gen_random_bytes(16), 'hex')), $5, $6, $7, now())
RETURNING *;

-- name: UsersUpdate :one
UPDATE users
SET
  name = $2,
  is_admin = $3,
  created_by = $4,
  api_token = $5,
  telegram_id = $6,
  lark_id = $7,
  slack_id = $8
WHERE id = $1
RETURNING *;

-- name: UsersList :many
SELECT * FROM users;

-- name: UsersGetByID :one
SELECT * FROM users WHERE id = $1;

-- name: UsersGetByName :one
SELECT * FROM users WHERE name = $1;

-- name: UsersGetByAPIToken :one
SELECT * FROM users WHERE api_token = @token::text;

-- name: UsersGetByTelegramID :one
SELECT * FROM users WHERE telegram_id = @id::bigint;

-- name: UsersGetByLarkID :one
SELECT * FROM users WHERE lark_id = @id::text;

-- name: UsersGetBySlackID :one
SELECT * FROM users WHERE slack_id = @id::text;

-- name: UsersDelete :exec
DELETE FROM users WHERE id = $1;
