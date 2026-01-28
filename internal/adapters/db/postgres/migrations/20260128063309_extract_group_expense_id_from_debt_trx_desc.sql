-- +goose Up
-- +goose StatementBegin
UPDATE debt_transactions
SET group_expense_id = (
    regexp_match(description, 'Share for group expense ([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})', 'i')
)[1]::uuid
WHERE description ~ 'Share for group expense [0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'
  AND group_expense_id IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
