CREATE extension pgcrypto;
UPDATE users SET params = params || jsonb_build_object('api.token', encode(gen_random_bytes(16), 'hex')) WHERE NOT params ? 'api.token' OR params->'api.token' IS NULL;
