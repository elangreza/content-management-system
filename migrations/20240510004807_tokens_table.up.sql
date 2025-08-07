BEGIN
;

CREATE TABLE IF NOT EXISTS "tokens" (
    "id" UUID PRIMARY KEY,
    "user_id" UUID NOT NULL REFERENCES "users" ("id"),
    "token" VARCHAR NOT NULL,
    "token_type" VARCHAR NOT NULL,
    "issued_at" TIMESTAMPTZ NOT NULL,
    "expired_at" TIMESTAMPTZ NOT NULL,
    "duration" VARCHAR NOT NULL
);

COMMIT;