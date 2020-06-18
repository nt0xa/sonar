WITH admins AS (
    SELECT * FROM users WHERE is_admin = true
)
UPDATE users SET params = jsonb_set(users.params, '{admin}', 'true') FROM admins WHERE users.id = admins.id;

ALTER TABLE users DROP COLUMN is_admin;
