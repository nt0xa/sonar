CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    payload_id INT NOT NULL REFERENCES payloads (id) ON DELETE CASCADE,
    protocol TEXT NOT NULL,
    rw BYTEA NOT NULL,
    r BYTEA NOT NULL,
    w BYTEA NOT NULL,
    meta JSONB NOT NULL,
    remote_addr TEXT NOT NULL,
    received_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);
