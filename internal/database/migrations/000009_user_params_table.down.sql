ALTER TABLE users ADD COLUMN params JSONB NOT NULL DEFAULT '{}';

UPDATE users
SET params = jsonb_set(users.params, '{telegram.id}', to_jsonb(p.value::INT))
FROM (
    SELECT user_id, key, value FROM user_params
) AS p
WHERE user_id = users.id AND p.key = 'telegram.id';

UPDATE users
SET params = jsonb_set(users.params, '{api.token}', to_jsonb(p.value::TEXT))
FROM (
    SELECT user_id, key, value FROM user_params
) AS p
WHERE user_id = users.id AND p.key = 'api.token';

DROP TABLE user_params;
