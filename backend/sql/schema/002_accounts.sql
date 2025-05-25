-- +goose Up
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    name TEXT NOT NULL,
    currency TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE accounts;