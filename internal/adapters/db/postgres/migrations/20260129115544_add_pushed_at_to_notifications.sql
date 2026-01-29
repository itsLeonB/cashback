-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS notifications
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS pushed_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS notifications
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS pushed_at;
-- +goose StatementEnd
