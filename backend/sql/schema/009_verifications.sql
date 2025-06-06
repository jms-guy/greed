-- +goose Up
CREATE TABLE verification_records (
    user_id UUID PRIMARY KEY REFERENCES users(id)
    ON DELETE CASCADE,
    verification_code TEXT NOT NULL,
    expiry_time TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE verification_records;