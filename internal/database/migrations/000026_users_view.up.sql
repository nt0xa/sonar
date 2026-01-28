CREATE VIEW users_full AS
SELECT
  users.*,
  COALESCE(json_object_agg(user_params.key, user_params.value)
           FILTER (WHERE user_params.key IS NOT NULL), '{}')::jsonb AS params
FROM users
LEFT JOIN user_params ON user_params.user_id = users.id
GROUP BY users.id;
