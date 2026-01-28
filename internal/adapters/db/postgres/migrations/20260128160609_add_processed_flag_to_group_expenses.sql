-- +goose Up
-- +goose StatementBegin
ALTER TABLE group_expenses
ADD COLUMN processed BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE group_expenses
DROP COLUMN processed;
-- +goose StatementEnd
