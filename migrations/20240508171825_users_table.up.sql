BEGIN
;

CREATE TABLE IF NOT EXISTS "users" (
    "id" UUID PRIMARY KEY,
    "name" VARCHAR(50) NOT NULL UNIQUE,
    "email" VARCHAR(50) NOT NULL UNIQUE,
    "password" BYTEA NOT NULL,
    "permission" BIGINT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NULL
);

CREATE INDEX "users_email_index" ON "users" ("email");

CREATE TRIGGER "log_user_update" BEFORE
UPDATE
    ON "users" FOR EACH ROW EXECUTE PROCEDURE log_update_master();

COMMIT;