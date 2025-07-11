-- +goose Up
ALTER TABLE users
ADD is_member BOOLEAN NOT NULL DEFAULT false;
-- +goose Down
ALTER TABLE users
DROP COLUMN is_member;