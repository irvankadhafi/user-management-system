-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS "resources" (
    id TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS "actions" (
    id TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS "role_resource_actions" (
                                                       "role" user_role NOT NULL,
                                                       "resource" TEXT NOT NULL REFERENCES "resources"(id),
    "action" TEXT NOT NULL REFERENCES "actions"(id)
    );

CREATE UNIQUE INDEX rra_unique_idx ON "role_resource_actions"("role", "resource", "action");

-- +migrate Down
DROP TABLE IF EXISTS "role_resource_actions";
DROP TABLE IF EXISTS "actions";
DROP TABLE IF EXISTS "resources";