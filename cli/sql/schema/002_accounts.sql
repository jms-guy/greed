-- +goose Up
CREATE TABLE accounts (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    subtype TEXT,
    mask TEXT,
    official_name TEXT,
    available_balance NUMERIC,
    current_balance NUMERIC,
    iso_currency_code TEXT,
    institution_name TEXT
);

-- +goose Down
DROP TABLE accounts;