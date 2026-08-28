-- +goose Up
-- +goose StatementBegin
ALTER TABLE debt_transactions
ADD CONSTRAINT no_self_debt CHECK (lender_profile_id <> borrower_profile_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE debt_transactions
DROP CONSTRAINT no_self_debt;
-- +goose StatementEnd
