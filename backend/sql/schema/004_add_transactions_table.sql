-- +goose Up
CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    amount NUMERIC (16, 2) NOT NULL,
    category TEXT NOT NULL,
    description TEXT,
    transaction_date TIMESTAMPTZ NOT NULL,
    transaction_type TEXT NOT NULL CHECK (transaction_type in ('debit', 'credit', 'transfer')),
    currency_code TEXT NOT NULL REFERENCES supported_currencies(code),
    account_id TEXT NOT NULL REFERENCES accounts(id)
    ON DELETE CASCADE
);

CREATE INDEX idx_transactions_date ON transactions(transaction_date);

-- +goose Down
DROP TABLE transactions;