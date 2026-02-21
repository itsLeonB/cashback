-- +goose Up
-- +goose StatementBegin
INSERT INTO subscriptions (profile_id, plan_version_id)
SELECT up.id, pv.id
FROM user_profiles up
CROSS JOIN plan_versions pv
WHERE up.user_id IS NOT NULL
  AND pv.is_default IS TRUE
  AND NOT EXISTS (
    SELECT 1 
    FROM subscriptions s 
    WHERE s.profile_id = up.id 
      AND s.plan_version_id = pv.id
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Cannot be rollbacked, safe to retry (idempotent)
-- +goose StatementEnd
