-- +goose Up
-- +goose StatementBegin
ALTER TABLE password_reset_tokens
ADD COLUMN selector TEXT NOT NULL DEFAULT '',
ADD COLUMN verifier_hash TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_password_reset_tokens_selector ON password_reset_tokens (selector) WHERE selector != '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_password_reset_tokens_selector;

ALTER TABLE password_reset_tokens
DROP COLUMN IF EXISTS verifier_hash,
DROP COLUMN IF EXISTS selector;
-- +goose StatementEnd
