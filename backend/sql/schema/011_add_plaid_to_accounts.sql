-- +goose Up
ALTER TABLE accounts
ADD plaid_account_id TEXT UNIQUE NOT NULL DEFAULT 'unset';
ALTER TABLE accounts
ADD item_id UUID REFERENCES plaid_items(id)
ON DELETE CASCADE;

-- +goose Down
ALTER TABLE accounts
DROP COLUMN plaid_account_id;
ALTER TABLE accounts
DROP COLUMN item_id;