-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS subscription_payments (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    subscription_id UUID NOT NULL REFERENCES subscriptions(id),
    amount NUMERIC(20, 2) NOT NULL,
    currency TEXT NOT NULL,
    gateway TEXT NOT NULL,
    gateway_transaction_id TEXT NOT NULL,
    gateway_subscription_id TEXT NOT NULL,
    status TEXT NOT NULL,
    failure_reason TEXT,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    gateway_event_id TEXT,
    paid_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS subscription_payments_subscription_id_idx ON subscription_payments(subscription_id);
CREATE UNIQUE INDEX IF NOT EXISTS subscription_payments_gateway_unique_idx ON subscription_payments(gateway_transaction_id, gateway);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS subscription_payments_gateway_unique_idx;
DROP INDEX IF EXISTS subscription_payments_subscription_id_idx;
DROP TABLE IF EXISTS subscription_payments;
-- +goose StatementEnd
