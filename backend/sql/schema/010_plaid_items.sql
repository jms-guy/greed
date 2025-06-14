-- +goose Up
CREATE TABLE plaid_items (
    id TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id)
    ON DELETE CASCADE,
    access_token TEXT NOT NULL,
    institution_name TEXT NOT NULL,
    request_id TEXT,
    nickname TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE plaid_items;