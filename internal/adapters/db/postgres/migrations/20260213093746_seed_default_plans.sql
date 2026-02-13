-- +goose Up
-- +goose StatementBegin
WITH new_plan AS (
    INSERT INTO plans (name, is_active)
    VALUES ('Free', true)
    RETURNING id
)
INSERT INTO plan_versions (plan_id, price_amount, price_currency, billing_interval, bill_uploads_daily, bill_uploads_monthly, effective_from, is_default)
SELECT id, 0, 'IDR', 'monthly', 0, 0, CURRENT_TIMESTAMP, true
FROM new_plan;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE TABLE plan_versions CASCADE;
TRUNCATE TABLE plans CASCADE;
-- +goose StatementEnd
