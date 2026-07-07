-- +goose Up
-- +goose StatementBegin
ALTER TABLE group_expense_other_fee_participants
DROP CONSTRAINT group_expense_other_fee_participants_other_fee_id_fkey;

ALTER TABLE group_expense_other_fee_participants
ADD CONSTRAINT group_expense_other_fee_participants_other_fee_id_fkey
FOREIGN KEY (other_fee_id) REFERENCES group_expense_other_fees(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE group_expense_other_fee_participants
DROP CONSTRAINT group_expense_other_fee_participants_other_fee_id_fkey;

ALTER TABLE group_expense_other_fee_participants
ADD CONSTRAINT group_expense_other_fee_participants_other_fee_id_fkey
FOREIGN KEY (other_fee_id) REFERENCES group_expense_other_fees(id);
-- +goose StatementEnd
