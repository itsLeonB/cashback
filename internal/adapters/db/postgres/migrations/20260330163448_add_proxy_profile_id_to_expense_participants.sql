-- +goose Up
-- +goose StatementBegin
ALTER TABLE group_expense_participants
ADD COLUMN IF NOT EXISTS proxy_profile_id UUID REFERENCES user_profiles(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE group_expense_participants
DROP COLUMN IF EXISTS proxy_profile_id;
-- +goose StatementEnd
