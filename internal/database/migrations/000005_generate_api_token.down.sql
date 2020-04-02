DROP extension IF EXISTS pgcrypto;
UPDATE users SET params = params #- '{apiToken}';
