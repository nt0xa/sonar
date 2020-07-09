CREATE TABLE IF NOT EXISTS dns_records (
    id SERIAL,
    payload_id INT NOT NULL REFERENCES payloads (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    ttl INT NOT NULL,
    values TEXT[] NOT NULL,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (payload_id, name, type)
);
