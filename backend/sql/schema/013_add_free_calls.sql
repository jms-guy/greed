-- +goose Up
ALTER TABLE users
ADD free_calls INT NOT NULL DEFAULT 10;

-- +goose Down
ALTER TABLE users
DROP COLUMN free_calls;