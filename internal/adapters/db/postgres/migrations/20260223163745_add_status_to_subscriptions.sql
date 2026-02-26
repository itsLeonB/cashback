-- +goose Up
-- +goose StatementBegin
ALTER TABLE subscriptions
ADD COLUMN IF NOT EXISTS status TEXT;
UPDATE subscriptions SET status = 'active';
ALTER TABLE subscriptions
ALTER COLUMN status SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE subscriptions
DROP COLUMN IF EXISTS status;
-- +goose StatementEnd
