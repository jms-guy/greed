-- +goose Up
ALTER TABLE users
ADD email TEXT NOT NULL DEFAULT 'unset';
ALTER TABLE users
ADD is_verified BOOLEAN DEFAULT false;

-- +goose Down
ALTER TABLE users
DROP COLUMN email;
ALTER TABLE users
DROP COLUMN is_verified;