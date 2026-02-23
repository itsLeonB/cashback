-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS group_expense_bills_created_at_idx ON group_expense_bills (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS group_expense_bills_created_at_idx;
-- +goose StatementEnd
