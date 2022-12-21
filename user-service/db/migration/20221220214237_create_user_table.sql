-- +migrate Up notransaction
CREATE TYPE "user_role" AS ENUM (
    'ADMIN',
    'MEMBER'
);


CREATE TABLE IF NOT EXISTS "users" (
    "id" uuid PRIMARY KEY,
    "name" text,
    "email" text,
    "password" text,
    "role" user_role,
    "phone_number" text,
    "created_by" uuid,
    "updated_by" uuid,
    "created_at" timestamp NOT NULL DEFAULT 'now()',
    "updated_at" timestamp NOT NULL DEFAULT 'now()',
    "deleted_at" timestamp
);

ALTER TABLE "users" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "users" ADD FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "users" ADD CONSTRAINT email_unique UNIQUE("email");

-- +migrate Down
DROP TABLE IF EXISTS "users";