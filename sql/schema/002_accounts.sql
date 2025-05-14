-- +goose Up
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    balance NUMERIC (10, 2),
    goal NUMERIC (10, 2),
    currency TEXT CHECK (currency IN ('CAD')),
    user_id UUID NOT NULL REFERENCES users(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE accounts;