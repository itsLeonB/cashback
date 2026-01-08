-- +goose Up
-- +goose StatementBegin
ALTER TABLE group_expense_other_fees DROP COLUMN deleted_at;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE group_expense_other_fees ADD COLUMN deleted_at TIMESTAMPTZ;
-- +goose StatementEnd
