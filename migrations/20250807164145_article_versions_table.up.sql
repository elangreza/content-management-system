BEGIN
;

CREATE TABLE IF NOT EXISTS "article_versions" (
    "id" SERIAL PRIMARY KEY,
    "article_id" INT NOT NULL REFERENCES articles("id"),
    "title" VARCHAR(255) NOT NULL,
    "body" TEXT NOT NULL,
    "version" INT NOT NULL,
    -- DRAFT 0, PUBLISHED 1, ARCHIVED 2
    "status" INT DEFAULT 0,
    "created_by" UUID NOT NULL REFERENCES users("id"),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_by" UUID REFERENCES users("id"),
    "updated_at" TIMESTAMPTZ NULL
);

CREATE TRIGGER "log_article_version_update" BEFORE
UPDATE
    ON "article_versions" FOR EACH ROW EXECUTE PROCEDURE log_update_master();

COMMIT;