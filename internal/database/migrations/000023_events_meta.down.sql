BEGIN TRANSACTION;

UPDATE events
SET meta = (meta - 'dns') || jsonb_build_object(
    'answer', meta->'dns'->'answer',
    'question', meta->'dns'->'question'
)
WHERE protocol = 'dns'
AND meta ? 'dns';

UPDATE events
SET meta = (meta - 'http') || jsonb_build_object(
    'request', meta->'http'->'request',
    'response', meta->'http'->'response'
)
WHERE protocol = 'http'
AND meta ? 'http';

UPDATE events
SET meta = (meta - 'smtp') || jsonb_build_object(
    'session', meta->'smtp'->'session',
    'email', meta->'smtp'->'email'
)
WHERE protocol = 'smtp'
AND meta ? 'smtp';

UPDATE events
SET meta = (meta - 'ftp') || jsonb_build_object(
    'session', meta->'ftp'->'session'
)
WHERE protocol = 'ftp'
AND meta ? 'ftp';

COMMIT;
