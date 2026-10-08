CREATE SEQUENCE information_asset_id_seq
    AS INTEGER
    MINVALUE 1
    MAXVALUE 99999
    START WITH 1
    INCREMENT BY 1
    NO CYCLE;

SELECT setval(
    'information_asset_id_seq',
    COALESCE(
        (SELECT MAX(SUBSTRING(id FROM 3)::INTEGER) FROM information_assets),
        1
    ),
    EXISTS (SELECT 1 FROM information_assets)
);
