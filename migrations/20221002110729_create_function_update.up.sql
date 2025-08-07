BEGIN
;

CREATE
OR REPLACE FUNCTION log_update_master() RETURNS TRIGGER AS $$ BEGIN
    NEW .updated_at = NOW();

RETURN NEW;

END;

$$ LANGUAGE 'plpgsql';

COMMIT;