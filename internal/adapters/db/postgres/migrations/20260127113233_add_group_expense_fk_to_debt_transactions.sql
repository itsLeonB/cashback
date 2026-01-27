-- +goose Up
-- +goose StatementBegin
ALTER TABLE debt_transactions
ADD COLUMN IF NOT EXISTS group_expense_id UUID;
CREATE INDEX IF NOT EXISTS debt_transactions_group_expense_id_idx ON debt_transactions(group_expense_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE debt_transactions
DROP COLUMN IF EXISTS group_expense_id;
DROP INDEX IF EXISTS debt_transactions_group_expense_id_idx;
-- +goose StatementEnd
