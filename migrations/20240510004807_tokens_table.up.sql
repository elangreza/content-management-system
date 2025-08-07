BEGIN
;

CREATE TABLE IF NOT EXISTS "tokens" (
    "id" UUID PRIMARY KEY,
    "user_id" UUID NOT NULL,
    "token" VARCHAR NOT NULL,
    "token_type" VARCHAR NOT NULL,
    "issued_at" TIMESTAMPTZ NOT NULL,
    "expired_at" TIMESTAMPTZ NOT NULL,
    "duration" VARCHAR NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NULL
);

CREATE INDEX "tokens_id_index" ON "tokens" ("id");

ALTER TABLE
    "tokens"
ADD
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON
DELETE
    CASCADE;

CREATE TRIGGER "log_token_update" BEFORE
UPDATE
    ON "tokens" FOR EACH ROW EXECUTE PROCEDURE log_update_master();

COMMIT;