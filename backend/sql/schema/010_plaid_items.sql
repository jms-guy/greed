-- +goose Up
CREATE TABLE plaid_items (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id)
    ON DELETE CASCADE,
    item_id TEXT UNIQUE NOT NULL,
    access_token TEXT NOT NULL,
    request_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE plaid_items;