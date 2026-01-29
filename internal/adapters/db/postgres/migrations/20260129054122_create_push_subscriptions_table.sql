-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuidv7(),
    profile_id UUID NOT NULL REFERENCES user_profiles(id),
    endpoint TEXT NOT NULL UNIQUE,
    keys JSONB NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS push_subscriptions_profile_id_idx ON push_subscriptions (profile_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS push_subscriptions;
DROP INDEX IF EXISTS push_subscriptions_profile_id_idx;
-- +goose StatementEnd
