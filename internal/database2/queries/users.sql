-- name: usersInsert :one
INSERT INTO users (name, is_admin, created_by, created_at)
VALUES ($1, $2, $3, now())
RETURNING *;

-- name: userParamsInsert :exec
INSERT INTO user_params(user_id, key, value)
VALUES ($1, $2, $3);

-- name: usersUpdate :one
UPDATE users
SET
  name = $2,
  is_admin = $3,
  created_by = $4
WHERE id = $1
RETURNING *;

-- name: userParamsUpdate :exec
UPDATE user_params
SET
  value = $3
WHERE user_id = $1 AND key = $2;

-- name: UsersList :many
SELECT * FROM users_full;

-- name: UsersGetByID :one
SELECT * FROM users_full WHERE id = $1;

-- name: UsersGetByName :one
SELECT * FROM users_full WHERE name = $1;

-- name: UsersGetByParam :one
SELECT * FROM users_full WHERE params->>sqlc.arg(key)::text = sqlc.arg(value)::text;

-- name: UsersDelete :exec
DELETE FROM users WHERE id = $1;
