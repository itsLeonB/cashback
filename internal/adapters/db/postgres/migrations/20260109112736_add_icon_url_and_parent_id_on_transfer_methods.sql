-- +goose Up
-- +goose StatementBegin
ALTER TABLE transfer_methods
ADD COLUMN icon_url TEXT,
ADD COLUMN parent_id UUID,
ADD COLUMN count BIGINT DEFAULT 0,
ADD CONSTRAINT transfer_methods_name_unique UNIQUE (name);

CREATE INDEX IF NOT EXISTS transfer_methods_parent_id_idx ON transfer_methods(parent_id) WHERE parent_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transfer_methods
DROP COLUMN icon_url,
DROP COLUMN parent_id,
DROP COLUMN count,
DROP CONSTRAINT IF EXISTS transfer_methods_name_unique;

DROP INDEX IF EXISTS transfer_methods_parent_id_idx;
-- +goose StatementEnd
