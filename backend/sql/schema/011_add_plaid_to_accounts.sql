-- +goose Up
ALTER TABLE accounts
ADD item_id TEXT NOT NULL REFERENCES plaid_items(id)
ON DELETE CASCADE;
ALTER TABLE accounts
ADD user_id UUID NOT NULL REFERENCES users(id)
ON DELETE CASCADE;

-- +goose Down
ALTER TABLE accounts
DROP COLUMN item_id;
ALTER TABLE accounts
DROP COLUMN user_id;