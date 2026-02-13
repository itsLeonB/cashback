-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS plan_versions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    plan_id UUID NOT NULL REFERENCES plans(id),
    price_amount NUMERIC(20, 2) NOT NULL,
    price_currency VARCHAR(3) NOT NULL,
    billing_interval TEXT NOT NULL,
    bill_uploads_daily SMALLINT NOT NULL,
    bill_uploads_monthly SMALLINT NOT NULL,
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    is_default BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    profile_id UUID NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    plan_version_id UUID NOT NULL REFERENCES plan_versions(id),
    ends_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    auto_renew BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS plan_versions_plan_id_idx ON plan_versions(plan_id);
CREATE INDEX IF NOT EXISTS subscriptions_profile_id_idx ON subscriptions(profile_id);
CREATE INDEX IF NOT EXISTS subscriptions_plan_version_id_idx ON subscriptions(plan_version_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS subscriptions_plan_version_id_idx;
DROP INDEX IF EXISTS subscriptions_profile_id_idx;
DROP INDEX IF EXISTS plan_versions_plan_id_idx;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plan_versions;
DROP TABLE IF EXISTS plans;
-- +goose StatementEnd
