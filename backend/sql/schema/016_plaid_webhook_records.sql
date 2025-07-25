-- +goose Up
CREATE TABLE plaid_webhook_records (
    id UUID PRIMARY KEY,
    webhook_type TEXT NOT NULL,
    webhook_code TEXT NOT NULL,
    item_id TEXT NOT NULL REFERENCES plaid_items(id)
    ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id)
    ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed BOOLEAN NOT NULL DEFAULT FALSE,
    processed_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE plaid_webhook_records;