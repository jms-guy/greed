-- +goose Up
CREATE TABLE recurring_streams (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(id)
    ON DELETE CASCADE, 
    description TEXT NOT NULL,
    merchant_name TEXT,
    frequency TEXT NOT NULL,
    is_active BOOLEAN NOT NULL,
    predicted_next_date TEXT,
    stream_type TEXT NOT NULL,
    created_at TIMESTAMP
);

CREATE TABLE transactions_to_streams (
    transaction_id TEXT REFERENCES transactions(id) ON DELETE CASCADE,
    stream_id TEXT REFERENCES recurring_streams(id) ON DELETE CASCADE,
    PRIMARY KEY (transaction_id, stream_id)
);

-- +goose Down
DROP TABLE transactions_to_streams;
DROP TABLE recurring_streams;