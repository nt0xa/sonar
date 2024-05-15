CREATE TYPE dns_record_type AS ENUM('A', 'AAAA', 'MX', 'TXT', 'CNAME');

CREATE TYPE dns_strategy AS ENUM('all', 'round-robin', 'rebind');

CREATE TABLE dns_records (
    id SERIAL,
    payload_id INT NOT NULL REFERENCES payloads (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type dns_record_type NOT NULL,
    ttl INT NOT NULL,
    values TEXT[] NOT NULL,
    strategy dns_strategy NOT NULL,
    last_answer TEXT[],
    last_accessed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (payload_id, name, type),
    CHECK (cardinality(values) > 0),
    CHECK (cardinality(last_answer) > 0)
);
