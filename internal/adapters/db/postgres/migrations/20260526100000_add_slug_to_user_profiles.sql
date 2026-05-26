-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_profiles
ADD COLUMN IF NOT EXISTS slug TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_profiles_slug ON user_profiles (slug) WHERE slug IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_profiles_slug;

ALTER TABLE user_profiles
DROP COLUMN IF EXISTS slug;
-- +goose StatementEnd
