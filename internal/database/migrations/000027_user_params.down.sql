BEGIN;

CREATE TABLE user_params (
    id bigint NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id INT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE (key, value)
);

INSERT INTO user_params (user_id, key, value)
SELECT u.id, 'telegram.id', u.telegram_id::text
FROM users u
WHERE u.telegram_id IS NOT NULL;

INSERT INTO user_params (user_id, key, value)
SELECT u.id, 'slack.id', u.slack_id
FROM users u
WHERE u.slack_id IS NOT NULL;

INSERT INTO user_params (user_id, key, value)
SELECT u.id, 'lark.userid', u.lark_id
FROM users u
WHERE u.lark_id IS NOT NULL;

INSERT INTO user_params (user_id, key, value)
SELECT u.id, 'api.token', u.api_token
FROM users u
WHERE u.api_token IS NOT NULL;

DROP INDEX users_telegram_id_unique;
DROP INDEX users_slack_id_unique;
DROP INDEX users_lark_id_unique;
DROP INDEX users_api_token_unique;

ALTER TABLE users
  DROP COLUMN telegram_id,
  DROP COLUMN slack_id,
  DROP COLUMN lark_id,
  DROP COLUMN api_token;

CREATE VIEW users_full AS
SELECT
  users.*,
  COALESCE(json_object_agg(user_params.key, user_params.value)
           FILTER (WHERE user_params.key IS NOT NULL), '{}')::jsonb AS params
FROM users
LEFT JOIN user_params ON user_params.user_id = users.id
GROUP BY users.id;

COMMIT;
