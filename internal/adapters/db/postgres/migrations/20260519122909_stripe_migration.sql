-- +goose Up
ALTER TABLE plan_versions ADD COLUMN stripe_price_id TEXT;
ALTER TABLE user_profiles ADD COLUMN stripe_customer_id TEXT;
CREATE INDEX idx_user_profiles_stripe_customer_id ON user_profiles(stripe_customer_id);
ALTER TABLE subscription_payments ALTER COLUMN gateway SET DEFAULT 'stripe';
ALTER TABLE subscriptions ADD COLUMN gateway_subscription_id TEXT;

-- +goose Down
ALTER TABLE subscriptions DROP COLUMN IF EXISTS gateway_subscription_id;
ALTER TABLE subscription_payments ALTER COLUMN gateway DROP DEFAULT;
DROP INDEX IF EXISTS idx_user_profiles_stripe_customer_id;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS stripe_customer_id;
ALTER TABLE plan_versions DROP COLUMN IF EXISTS stripe_price_id;
