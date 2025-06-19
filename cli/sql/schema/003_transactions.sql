-- +goose Up
CREATE TABLE transactions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(id)
    ON DELETE CASCADE,
    amount NUMERIC NOT NULL,
    iso_currency_code TEXT,
    date TEXT,
    merchant_name TEXT,
    payment_channel TEXT NOT NULL,
    personal_finance_category TEXT NOT NULL,
    created_at TEXT,
    updated_at TEXT
);

-- +goose Down
DROP TABLE transactions;