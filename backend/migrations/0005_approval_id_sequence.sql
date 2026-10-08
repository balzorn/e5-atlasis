CREATE SEQUENCE approval_id_seq
    AS INTEGER
    MINVALUE 1 MAXVALUE 99999 START WITH 1 INCREMENT BY 1 NO CYCLE;

SELECT setval(
    'approval_id_seq',
    COALESCE(
        (SELECT MAX(SUBSTRING(id FROM 4)::INTEGER) FROM approvals WHERE id ~ '^APR[0-9]{5}$'),
        1
    ),
    EXISTS (SELECT 1 FROM approvals WHERE id ~ '^APR[0-9]{5}$')
);
