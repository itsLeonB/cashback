-- +goose Up
-- +goose StatementBegin
ALTER TABLE subscriptions
ADD COLUMN IF NOT EXISTS current_period_start TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS current_period_end TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE subscriptions
DROP COLUMN IF EXISTS current_period_start,
DROP COLUMN IF EXISTS current_period_end;
-- +goose StatementEnd
