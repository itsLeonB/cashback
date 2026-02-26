-- +goose Up
-- +goose StatementBegin
ALTER TABLE subscription_payments ADD COLUMN expired_at TIMESTAMPTZ;
CREATE UNIQUE INDEX unique_incomplete_payment_idx ON subscription_payments (subscription_id) WHERE status IN ('pending', 'processing');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX unique_incomplete_payment_idx;
ALTER TABLE subscription_payments DROP COLUMN expired_at;
-- +goose StatementEnd
