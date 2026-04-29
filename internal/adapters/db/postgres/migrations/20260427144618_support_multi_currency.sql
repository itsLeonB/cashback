-- +goose Up
-- +goose StatementBegin
ALTER TABLE group_expenses
ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
ADD COLUMN IF NOT EXISTS fx_rate_to_home_currency NUMERIC(20, 6),
ADD COLUMN IF NOT EXISTS fx_home_currency VARCHAR(3),
ADD COLUMN IF NOT EXISTS fx_locked_at TIMESTAMPTZ;

ALTER TABLE debt_transactions
ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'IDR';

ALTER TABLE user_profiles
ADD COLUMN IF NOT EXISTS home_currency VARCHAR(3) NOT NULL DEFAULT 'IDR';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE group_expenses
DROP COLUMN IF EXISTS currency,
DROP COLUMN IF EXISTS fx_rate_to_home_currency,
DROP COLUMN IF EXISTS fx_home_currency,
DROP COLUMN IF EXISTS fx_locked_at;

ALTER TABLE debt_transactions
DROP COLUMN IF EXISTS currency;

ALTER TABLE user_profiles
DROP COLUMN IF EXISTS home_currency;
-- +goose StatementEnd
