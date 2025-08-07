BEGIN
;

CREATE TABLE IF NOT EXISTS "articles" (
    "id" SERIAL PRIMARY KEY,
    "created_by" UUID NOT NULL REFERENCES users("id"),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_by" UUID REFERENCES users("id"),
    "updated_at" TIMESTAMPTZ NULL
);

CREATE TRIGGER "log_article_update" BEFORE
UPDATE
    ON "articles" FOR EACH ROW EXECUTE PROCEDURE log_update_master();

COMMIT;