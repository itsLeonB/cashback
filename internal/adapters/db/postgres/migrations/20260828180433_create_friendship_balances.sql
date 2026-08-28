-- +goose Up
-- +goose StatementBegin
CREATE TABLE friendship_balances (
    id uuid DEFAULT uuidv7() NOT NULL PRIMARY KEY,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    friendship_id uuid NOT NULL REFERENCES friendships(id) ON DELETE CASCADE,
    currency VARCHAR(3) NOT NULL,
    net_balance numeric(20,2) NOT NULL DEFAULT 0,
    CONSTRAINT unique_friendship_balance UNIQUE (friendship_id, currency)
);

CREATE INDEX friendship_balances_friendship_id_idx ON friendship_balances(friendship_id);

-- Backfill: pre-existing friendships have no cache row yet and only get one lazily on their
-- next debt_transactions write - without this, every existing friend pair would show a
-- zero/missing balance until then. net_balance is relative to friendships.profile_id1,
-- matching the app-level convention (positive = profile_id1 is net lender).
INSERT INTO friendship_balances (friendship_id, currency, net_balance)
SELECT f.id,
       dt.currency,
       SUM(CASE WHEN dt.lender_profile_id = f.profile_id1 THEN dt.amount ELSE -dt.amount END)
FROM friendships f
JOIN debt_transactions dt
  ON (dt.lender_profile_id = f.profile_id1 AND dt.borrower_profile_id = f.profile_id2)
  OR (dt.lender_profile_id = f.profile_id2 AND dt.borrower_profile_id = f.profile_id1)
GROUP BY f.id, dt.currency
HAVING SUM(CASE WHEN dt.lender_profile_id = f.profile_id1 THEN dt.amount ELSE -dt.amount END) <> 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS friendship_balances_friendship_id_idx;
DROP TABLE IF EXISTS friendship_balances;
-- +goose StatementEnd
