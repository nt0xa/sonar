UPDATE users SET params = params || jsonb_build_object('telegram.id', id) WHERE NOT params ? 'telegram.id' OR params->'telegram.id' IS NULL;
