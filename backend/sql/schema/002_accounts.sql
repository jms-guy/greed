-- +goose Up
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    balance NUMERIC (16, 2),
    goal NUMERIC (16, 2),
    currency TEXT NOT NULL,
    user_id UUID UNIQUE NOT NULL REFERENCES users(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE accounts;