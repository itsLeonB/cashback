-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuidv7(),
    profile_id UUID NOT NULL REFERENCES user_profiles(id),
    type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    metadata JSONB,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS notifications_profile_unread_created_idx
ON notifications (profile_id, created_at DESC)
WHERE read_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS notifications_unique_entity_idx
ON notifications (profile_id, type, entity_type, entity_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notifications;
DROP INDEX IF EXISTS notifications_profile_unread_created_idx;
DROP INDEX IF EXISTS notifications_unique_entity_idx;
-- +goose StatementEnd
