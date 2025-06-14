-- +goose Up
CREATE TABLE transactions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(id)
    ON DELETE CASCADE,
    amount NUMERIC(16, 2) NOT NULL,
    iso_currency_code TEXT,
    date TIMESTAMPTZ,
    merchant_name TEXT,
    payment_channel TEXT NOT NULL DEFAULT 'unset',
    personal_finance_category TEXT NOT NULL DEFAULT 'unset',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_date ON transactions(date);

-- +goose Down
DROP TABLE transactions;