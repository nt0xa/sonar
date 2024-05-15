CREATE TABLE user_params (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE (key, value)
);

INSERT INTO user_params
SELECT nextval('user_params_id_seq'), id, (jsonb_each_text(params)).* FROM users;
ALTER TABLE users DROP COLUMN params;
