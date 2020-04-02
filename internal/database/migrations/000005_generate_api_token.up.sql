CREATE extension IF NOT EXISTS pgcrypto;
UPDATE users SET params = params || jsonb_build_object('apiToken', encode(gen_random_bytes(16), 'hex')) WHERE NOT params ? 'apiToken' OR params->'apiToken' IS NULL;
