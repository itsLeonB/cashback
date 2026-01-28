-- +goose Up
-- +goose StatementBegin
ALTER TABLE debt_transactions DROP COLUMN IF EXISTS type;
ALTER TABLE debt_transactions DROP COLUMN IF EXISTS action;
DROP TYPE IF EXISTS debt_transaction_type;
DROP TYPE IF EXISTS debt_transaction_action;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TYPE debt_transaction_type AS ENUM ('LEND', 'REPAY');
CREATE TYPE debt_transaction_action AS ENUM ('LEND', 'BORROW', 'RECEIVE', 'RETURN');
ALTER TABLE debt_transactions ADD COLUMN type debt_transaction_type NOT NULL DEFAULT 'LEND';
ALTER TABLE debt_transactions ADD COLUMN action debt_transaction_action NOT NULL DEFAULT 'LEND';
-- +goose StatementEnd
