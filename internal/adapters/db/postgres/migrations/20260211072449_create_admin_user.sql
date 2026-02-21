-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS admin_roles (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS admin_users_roles (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES admin_roles(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS admin_users_roles_user_id_idx ON admin_users_roles(user_id);
CREATE INDEX IF NOT EXISTS admin_users_roles_role_id_idx ON admin_users_roles(role_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS admin_users_roles_role_id_idx;
DROP INDEX IF EXISTS admin_users_roles_user_id_idx;
DROP TABLE IF EXISTS admin_users_roles;
DROP TABLE IF EXISTS admin_roles;
DROP TABLE IF EXISTS admin_users;
-- +goose StatementEnd
