-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS push_subscriptions_profile_id_endpoint_idx;
CREATE UNIQUE INDEX IF NOT EXISTS push_subscriptions_profile_endpoint_unique_idx
ON push_subscriptions (profile_id, endpoint);
ALTER TABLE push_subscriptions
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS push_subscriptions_profile_endpoint_unique_idx;
CREATE INDEX IF NOT EXISTS push_subscriptions_profile_id_endpoint_idx ON push_subscriptions (profile_id, endpoint);
ALTER TABLE push_subscriptions
DROP COLUMN IF EXISTS updated_at;
-- +goose StatementEnd
