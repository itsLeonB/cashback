-- +goose Up
-- +goose StatementBegin
CREATE TABLE profile_transfer_methods (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    profile_id UUID NOT NULL REFERENCES user_profiles (id),
    transfer_method_id UUID NOT NULL REFERENCES transfer_methods (id),
    account_name TEXT NOT NULL,
    account_number TEXT NOT NULL
);

CREATE INDEX profile_transfer_methods_profile_id_idx ON profile_transfer_methods (profile_id);
CREATE INDEX profile_transfer_methods_transfer_method_id_idx ON profile_transfer_methods (transfer_method_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS profile_transfer_methods;
DROP INDEX profile_transfer_methods_profile_id_idx;
DROP INDEX profile_transfer_methods_transfer_method_id_idx;
-- +goose StatementEnd
