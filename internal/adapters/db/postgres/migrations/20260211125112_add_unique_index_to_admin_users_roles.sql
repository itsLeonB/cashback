-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS admin_users_roles_unique_idx ON admin_users_roles (user_id, role_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS admin_users_roles_unique_idx;
-- +goose StatementEnd
