BEGIN;

ALTER TABLE users
  ADD COLUMN telegram_id bigint,
  ADD COLUMN slack_id text,
  ADD COLUMN lark_id text,
  ADD COLUMN api_token text;

-- Backfill (assumes one row per (user_id, key))
UPDATE users u
SET telegram_id = NULLIF(BTRIM(up.value), '')::bigint
FROM user_params up
WHERE up.user_id = u.id
  AND up.key = 'telegram.id'
  AND u.telegram_id IS NULL;

UPDATE users u
SET slack_id = NULLIF(BTRIM(up.value), '')
FROM user_params up
WHERE up.user_id = u.id
  AND up.key = 'slack.id'
  AND u.slack_id IS NULL;

UPDATE users u
SET lark_id = NULLIF(BTRIM(up.value), '')
FROM user_params up
WHERE up.user_id = u.id
  AND up.key = 'lark.userid'
  AND u.lark_id IS NULL;

UPDATE users u
SET api_token = NULLIF(BTRIM(up.value), '')
FROM user_params up
WHERE up.user_id = u.id
  AND up.key = 'api.token'
  AND u.api_token IS NULL;

-- Unique when NOT NULL (allows many NULLs)
CREATE UNIQUE INDEX users_telegram_id_unique
  ON users (telegram_id)
  WHERE telegram_id IS NOT NULL;

CREATE UNIQUE INDEX users_slack_id_unique
  ON users (slack_id)
  WHERE slack_id IS NOT NULL;

CREATE UNIQUE INDEX users_lark_id_unique
  ON users (lark_id)
  WHERE lark_id IS NOT NULL;

CREATE UNIQUE INDEX users_api_token_unique
  ON users (api_token)
  WHERE api_token IS NOT NULL;

-- Drop user_params after copying
DROP VIEW users_full;
DROP TABLE user_params;

COMMIT;


