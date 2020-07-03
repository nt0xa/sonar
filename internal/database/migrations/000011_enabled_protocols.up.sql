ALTER TABLE payloads RENAME COLUMN handlers TO notify_protocols;
UPDATE payloads SET notify_protocols = '{dns,http,smtp}';
