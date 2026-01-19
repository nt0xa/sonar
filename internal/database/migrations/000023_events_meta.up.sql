BEGIN TRANSACTION;

UPDATE events
SET meta = (meta - 'answer' - 'question') || jsonb_build_object(
    'dns', jsonb_build_object(
        'answer', meta->'answer',
        'question', meta->'question'
    )
)
WHERE protocol = 'dns'
AND meta ? 'answer'
AND NOT meta ? 'dns';

UPDATE events
SET meta = (meta - 'request' - 'response') || jsonb_build_object(
    'http', jsonb_build_object(
        'request', meta->'request',
        'response', meta->'response'
    )
)
WHERE protocol = 'http'
AND meta ? 'request'
AND NOT meta ? 'http';

UPDATE events
SET meta = (meta - 'session' - 'email') || jsonb_build_object(
    'smtp', jsonb_build_object(
        'session', meta->'session',
        'email', meta->'email'
    )
)
WHERE protocol = 'smtp'
AND meta ? 'email'
AND NOT meta ? 'smtp';

UPDATE events
SET meta = (meta - 'session') || jsonb_build_object(
    'ftp', jsonb_build_object(
        'session', meta->'session'
    )
)
WHERE protocol = 'ftp'
AND NOT meta ? 'ftp';

COMMIT;
