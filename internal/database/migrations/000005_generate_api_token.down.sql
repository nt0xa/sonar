DROP extension IF EXISTS pgcrypto;
UPDATE users SET params = params #- '{api.token}';
