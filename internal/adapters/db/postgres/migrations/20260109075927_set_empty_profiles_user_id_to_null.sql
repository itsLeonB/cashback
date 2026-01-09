-- +goose Up
-- +goose StatementBegin
UPDATE user_profiles
SET user_id = NULL
WHERE user_id = '00000000-0000-0000-0000-000000000000';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE user_profiles
SET user_id = '00000000-0000-0000-0000-000000000000'
WHERE user_id IS NULL;
-- +goose StatementEnd
