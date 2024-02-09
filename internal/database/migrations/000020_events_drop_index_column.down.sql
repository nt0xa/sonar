ALTER TABLE events ADD COLUMN index INT NOT NULL DEFAULT 0;
UPDATE events e SET index = (SELECT index FROM (SELECT id, ROW_NUMBER() OVER(PARTITION BY payload_id ORDER BY id ASC) AS index FROM events) q WHERE q.id = e.id);
ALTER TABLE events ADD UNIQUE (payload_id, index);
