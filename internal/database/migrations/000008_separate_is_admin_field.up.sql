ALTER TABLE users ADD COLUMN is_admin BOOL NOT NULL DEFAULT false;

WITH admins AS (
    SELECT * FROM users WHERE (params->>'admin')::BOOL = true
)
UPDATE users SET is_admin = true FROM admins WHERE users.id = admins.id;

UPDATE users SET params = params #- '{admin}';
