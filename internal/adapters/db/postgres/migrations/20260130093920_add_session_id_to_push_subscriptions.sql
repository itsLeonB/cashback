-- +goose Up
-- +goose StatementBegin
ALTER TABLE push_subscriptions 
ADD COLUMN session_id UUID REFERENCES sessions(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS push_subscriptions_session_id_idx ON push_subscriptions (session_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS push_subscriptions_session_id_idx;
ALTER TABLE push_subscriptions DROP COLUMN session_id;
-- +goose StatementEnd
