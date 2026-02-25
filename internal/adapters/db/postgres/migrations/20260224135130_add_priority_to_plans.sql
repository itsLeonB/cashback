-- +goose Up
-- +goose StatementBegin
ALTER TABLE plans
ADD COLUMN IF NOT EXISTS priority INT NOT NULL DEFAULT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE plans
DROP COLUMN IF EXISTS priority;
-- +goose StatementEnd
