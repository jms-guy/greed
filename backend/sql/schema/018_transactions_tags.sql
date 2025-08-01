-- +goose Up
CREATE TABLE transactions_to_tags (
    transaction_id TEXT NOT NULL REFERENCES transactions(id)
    ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES transaction_tags(id)
    ON DELETE CASCADE,
    PRIMARY KEY (transaction_id, tag_id)
);

-- +goose Down
DROP TABLE transactions_to_tags;