ALTER TABLE http_routes ADD COLUMN index INT NOT NULL DEFAULT 0;
UPDATE http_routes hr SET index = (SELECT index FROM (SELECT id, ROW_NUMBER() OVER(PARTITION BY payload_id ORDER BY id ASC) AS index FROM http_routes) q WHERE q.id = hr.id);
ALTER TABLE http_routes ADD UNIQUE (payload_id, index);

ALTER TABLE dns_records ADD COLUMN index INT NOT NULL DEFAULT 0;
UPDATE dns_records dr SET index = (SELECT index FROM (SELECT id, ROW_NUMBER() OVER(PARTITION BY payload_id ORDER BY id ASC) AS index FROM dns_records) q WHERE q.id = dr.id);
ALTER TABLE dns_records ADD UNIQUE (payload_id, index);

ALTER TABLE events ADD COLUMN index INT NOT NULL DEFAULT 0;
UPDATE events e SET index = (SELECT index FROM (SELECT id, ROW_NUMBER() OVER(PARTITION BY payload_id ORDER BY id ASC) AS index FROM events) q WHERE q.id = e.id);
ALTER TABLE events ADD UNIQUE (payload_id, index);
