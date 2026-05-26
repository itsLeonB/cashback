-- +goose Up
-- +goose StatementBegin
UPDATE user_profiles
SET slug = CONCAT(
    TRIM(BOTH '-' FROM REGEXP_REPLACE(LOWER(TRIM(name)), '[^a-z0-9]+', '-', 'g')),
    '-',
    LEFT(REPLACE(id::text, '-', ''), 6)
)
WHERE user_id IS NULL
  AND slug IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE user_profiles
SET slug = NULL
WHERE user_id IS NULL;
-- +goose StatementEnd
