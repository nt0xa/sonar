DELETE FROM dns_records WHERE type = 'NS'::dns_record_type;
ALTER TABLE dns_records ALTER type TYPE text;
ALTER TYPE dns_record_type RENAME TO dns_record_type_old;
CREATE TYPE dns_record_type AS ENUM('A', 'AAAA', 'MX', 'TXT', 'CNAME');
ALTER TABLE dns_records ALTER type TYPE dns_record_type USING type::dns_record_type;
DROP TYPE dns_record_type_old;
