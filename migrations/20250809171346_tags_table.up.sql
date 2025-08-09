BEGIN
;

CREATE TABLE IF NOT EXISTS "tags" (
    "name" VARCHAR PRIMARY KEY,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NULL
);

CREATE TRIGGER "log_tag_update" BEFORE
UPDATE
    ON "tags" FOR EACH ROW EXECUTE PROCEDURE log_update_master();

CREATE TABLE IF NOT EXISTS "article_version_tags" (
    "tag_name" VARCHAR REFERENCES tags(name),
    "article_version_id" BIGINT REFERENCES article_versions(id),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY ("tag_name", "article_version_id")
);

COMMIT;