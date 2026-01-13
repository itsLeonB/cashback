-- +goose Up
-- +goose StatementBegin
ALTER TABLE group_expense_item_participants
ADD COLUMN weight SMALLINT NOT NULL DEFAULT 1,
ADD COLUMN allocated_amount NUMERIC(20, 2) NOT NULL DEFAULT 0,
ALTER COLUMN share DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE group_expense_item_participants
DROP COLUMN weight,
DROP COLUMN allocated_amount,
ALTER COLUMN share SET NOT NULL;
-- +goose StatementEnd
