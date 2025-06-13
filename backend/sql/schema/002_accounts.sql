-- +goose Up
CREATE TABLE accounts (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    subtype TEXT,
    mask TEXT,
    official_name TEXT,
    available_balance NUMERIC (16, 2),
    current_balance NUMERIC (16, 2),
    iso_currency_code TEXT
);

-- +goose Down
DROP TABLE accounts;