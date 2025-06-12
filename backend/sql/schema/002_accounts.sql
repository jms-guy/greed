-- +goose Up
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    name TEXT,
    type TEXT,
    mask TEXT,
    official_name TEXT
);

-- +goose Down
DROP TABLE accounts;