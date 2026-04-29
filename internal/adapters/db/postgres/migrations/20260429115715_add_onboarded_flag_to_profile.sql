-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_profiles
ADD COLUMN IF NOT EXISTS onboarded_at TIMESTAMPTZ;

UPDATE user_profiles
SET onboarded_at = CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_profiles
DROP COLUMN IF EXISTS onboarded_at;
-- +goose StatementEnd
