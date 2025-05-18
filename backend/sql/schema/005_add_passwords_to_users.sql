-- +goose Up
ALTER TABLE users
ADD hashed_password TEXT NOT NULL DEFAULT 'unset';

-- +goose Down
ALTER TABLE users
DROP COLUMN hashed_password;