-- +goose Up
-- Enable Extensions
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create Types
CREATE TYPE friendship_type AS ENUM ('REAL', 'ANON');
CREATE TYPE debt_transaction_type AS ENUM ('LEND', 'REPAY');
CREATE TYPE debt_transaction_action AS ENUM ('LEND', 'BORROW', 'RECEIVE', 'RETURN');
CREATE TYPE fee_calculation_method AS ENUM ('EQUAL_SPLIT', 'ITEMIZED_SPLIT');

-- Create Tables
-- Users & Profiles (from Cocoon)
CREATE TABLE users (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    email text NOT NULL UNIQUE,
    password text NOT NULL,
    verified_at timestamp with time zone
);

CREATE TABLE user_profiles (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    user_id uuid,
    name text NOT NULL,
    avatar text
);
COMMENT ON COLUMN user_profiles.user_id IS 'Nullable. Can be NULL for peers who do not have an account in the app';

CREATE TABLE oauth_accounts (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id),
    provider text NOT NULL,
    provider_id text NOT NULL,
    email text,
    CONSTRAINT oauth_accounts_provider_provider_id_unique UNIQUE (provider, provider_id)
);

CREATE TABLE password_reset_tokens (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token character varying(255) NOT NULL UNIQUE,
    expires_at timestamp with time zone NOT NULL
);

CREATE TABLE related_profiles (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    real_profile_id uuid NOT NULL REFERENCES user_profiles(id),
    anon_profile_id uuid NOT NULL REFERENCES user_profiles(id) UNIQUE
);

CREATE TABLE friendships (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    profile_id1 uuid NOT NULL REFERENCES user_profiles(id),
    profile_id2 uuid NOT NULL REFERENCES user_profiles(id),
    type friendship_type NOT NULL,
    CONSTRAINT unique_friendship UNIQUE (profile_id1, profile_id2),
    CONSTRAINT profile_order CHECK (profile_id1 < profile_id2)
);

CREATE TABLE friendship_requests (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    sender_profile_id uuid NOT NULL REFERENCES user_profiles(id),
    recipient_profile_id uuid NOT NULL REFERENCES user_profiles(id),
    blocked_at timestamp with time zone
);

-- Transfer Methods & Debt Transactions (from Drex)
CREATE TABLE transfer_methods (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    name text NOT NULL,
    display text NOT NULL
);

CREATE TABLE debt_transactions (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    lender_profile_id uuid NOT NULL REFERENCES user_profiles(id),
    borrower_profile_id uuid NOT NULL REFERENCES user_profiles(id),
    type debt_transaction_type NOT NULL,
    action debt_transaction_action NOT NULL,
    amount numeric(20,2) NOT NULL,
    transfer_method_id uuid NOT NULL REFERENCES transfer_methods(id),
    description text
);

-- Group Expenses (from Billsplittr)
CREATE TABLE group_expenses (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    description text,
    status text DEFAULT 'DRAFT' NOT NULL,
    total_amount numeric(20,2) NOT NULL,
    items_total numeric(20,2) DEFAULT 0 NOT NULL,
    fees_total numeric(20,2) DEFAULT 0 NOT NULL,
    payer_profile_id uuid REFERENCES user_profiles(id),
    creator_profile_id uuid NOT NULL REFERENCES user_profiles(id)
);

CREATE TABLE group_expense_items (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    group_expense_id uuid NOT NULL REFERENCES group_expenses(id) ON DELETE CASCADE,
    name text NOT NULL,
    amount numeric(20,2) NOT NULL,
    quantity integer NOT NULL
);

CREATE TABLE group_expense_participants (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    group_expense_id uuid NOT NULL REFERENCES group_expenses(id) ON DELETE CASCADE,
    participant_profile_id uuid NOT NULL REFERENCES user_profiles(id),
    share_amount numeric(20,2) NOT NULL,
    CONSTRAINT unique_expense_profile UNIQUE (group_expense_id, participant_profile_id)
);

CREATE TABLE group_expense_item_participants (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expense_item_id uuid NOT NULL REFERENCES group_expense_items(id) ON DELETE CASCADE,
    profile_id uuid NOT NULL REFERENCES user_profiles(id),
    share numeric(20,4) NOT NULL,
    CONSTRAINT unique_expense_item_profile UNIQUE (expense_item_id, profile_id)
);

CREATE TABLE group_expense_bills (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    image_name text NOT NULL,
    group_expense_id uuid NOT NULL REFERENCES group_expenses(id) ON DELETE CASCADE,
    status text DEFAULT 'PENDING' NOT NULL,
    extracted_text text
);

CREATE TABLE group_expense_other_fees (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    group_expense_id uuid NOT NULL REFERENCES group_expenses(id) ON DELETE CASCADE,
    name text NOT NULL,
    amount numeric(20,2) NOT NULL,
    deleted_at timestamp with time zone,
    calculation_method fee_calculation_method
);

CREATE TABLE group_expense_other_fee_participants (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    other_fee_id uuid NOT NULL REFERENCES group_expense_other_fees(id),
    profile_id uuid NOT NULL REFERENCES user_profiles(id),
    share_amount numeric(20,2) NOT NULL,
    CONSTRAINT unique_fee_participant UNIQUE (other_fee_id, profile_id)
);

-- Indexes (deduplicated from PK creation)
CREATE INDEX user_profiles_user_id_idx ON user_profiles(user_id);
CREATE INDEX user_profiles_name_idx ON user_profiles(name);
CREATE INDEX friendships_profile_id1_idx ON friendships(profile_id1);
CREATE INDEX friendships_profile_id2_idx ON friendships(profile_id2);
CREATE INDEX friendships_type_idx ON friendships(type);
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
CREATE INDEX idx_user_profiles_name_trgm ON user_profiles USING gin (name gin_trgm_ops);
CREATE INDEX debt_transactions_lender_profile_id_idx ON debt_transactions(lender_profile_id);
CREATE INDEX debt_transactions_borrower_profile_id_idx ON debt_transactions(borrower_profile_id);
CREATE INDEX debt_transactions_transfer_method_id_idx ON debt_transactions(transfer_method_id);
CREATE INDEX debt_transactions_created_at_idx ON debt_transactions(created_at);
CREATE INDEX group_expenses_payer_profile_id_idx ON group_expenses(payer_profile_id);
CREATE INDEX group_expenses_created_at_idx ON group_expenses(created_at);
CREATE INDEX group_expense_participants_group_expense_id_idx ON group_expense_participants(group_expense_id);
CREATE INDEX group_expense_participants_participant_profile_id_idx ON group_expense_participants(participant_profile_id);
CREATE INDEX group_expense_participants_created_at_idx ON group_expense_participants(created_at);

-- +goose Down
DROP INDEX IF EXISTS group_expense_participants_created_at_idx;
DROP INDEX IF EXISTS group_expense_participants_participant_profile_id_idx;
DROP INDEX IF EXISTS group_expense_participants_group_expense_id_idx;
DROP INDEX IF EXISTS group_expenses_created_at_idx;
DROP INDEX IF EXISTS group_expenses_payer_profile_id_idx;
DROP INDEX IF EXISTS debt_transactions_created_at_idx;
DROP INDEX IF EXISTS debt_transactions_transfer_method_id_idx;
DROP INDEX IF EXISTS debt_transactions_borrower_profile_id_idx;
DROP INDEX IF EXISTS debt_transactions_lender_profile_id_idx;
DROP INDEX IF EXISTS idx_user_profiles_name_trgm;
DROP INDEX IF EXISTS idx_password_reset_tokens_expires_at;
DROP INDEX IF EXISTS idx_password_reset_tokens_user_id;
DROP INDEX IF EXISTS friendships_type_idx;
DROP INDEX IF EXISTS friendships_profile_id2_idx;
DROP INDEX IF EXISTS friendships_profile_id1_idx;
DROP INDEX IF EXISTS user_profiles_name_idx;
DROP INDEX IF EXISTS user_profiles_user_id_idx;

DROP TABLE IF EXISTS group_expense_other_fee_participants;
DROP TABLE IF EXISTS group_expense_other_fees;
DROP TABLE IF EXISTS group_expense_bills;
DROP TABLE IF EXISTS group_expense_item_participants;
DROP TABLE IF EXISTS group_expense_participants;
DROP TABLE IF EXISTS group_expense_items;
DROP TABLE IF EXISTS group_expenses;
DROP TABLE IF EXISTS debt_transactions;
DROP TABLE IF EXISTS transfer_methods;
DROP TABLE IF EXISTS friendship_requests;
DROP TABLE IF EXISTS friendships;
DROP TABLE IF EXISTS related_profiles;
DROP TABLE IF EXISTS profile_names;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS oauth_accounts;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS fee_calculation_method;
DROP TYPE IF EXISTS debt_transaction_action;
DROP TYPE IF EXISTS debt_transaction_type;
DROP TYPE IF EXISTS friendship_type;

DROP EXTENSION IF EXISTS pg_trgm;
