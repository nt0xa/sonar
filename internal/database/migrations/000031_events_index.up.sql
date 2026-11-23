BEGIN;

ALTER TABLE payloads ADD COLUMN last_event_index bigint NOT NULL DEFAULT 0;

ALTER TABLE events ADD COLUMN index bigint;
UPDATE events e SET index = n.index
FROM (SELECT id, ROW_NUMBER() OVER (PARTITION BY payload_id ORDER BY id) AS index FROM events) n
WHERE n.id = e.id;
ALTER TABLE events ALTER COLUMN index SET NOT NULL;
ALTER TABLE events ADD CONSTRAINT events_payload_id_index_key UNIQUE (payload_id, index);

UPDATE payloads p SET last_event_index = e.index
FROM (SELECT payload_id, MAX(index) AS index FROM events GROUP BY payload_id) e
WHERE e.payload_id = p.id;

COMMIT;
