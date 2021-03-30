CREATE TABLE IF NOT EXISTS http_routes (
    id SERIAL PRIMARY KEY,
    payload_id INT NOT NULL REFERENCES payloads (id) ON DELETE CASCADE,
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    code INT NOT NULL,
    headers JSONB NOT NULL,
    body BYTEA,
    is_dynamic BOOL NOT NULL,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (payload_id, method, path)
);
