CREATE SEQUENCE change_request_id_seq
    AS INTEGER
    MINVALUE 1
    MAXVALUE 99999
    START WITH 1
    INCREMENT BY 1
    NO CYCLE;

SELECT setval(
    'change_request_id_seq',
    COALESCE(MAX(SUBSTRING(id FROM 3)::INTEGER), 1),
    MAX(SUBSTRING(id FROM 3)::INTEGER) IS NOT NULL
)
FROM change_requests
WHERE id ~ '^CR[0-9]{5}$';
