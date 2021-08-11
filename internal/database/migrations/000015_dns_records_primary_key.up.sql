ALTER TABLE dns_records DROP CONSTRAINT IF EXISTS dns_records_pkey;
ALTER TABLE dns_records ADD CONSTRAINT dns_records_pkey PRIMARY KEY (id);
