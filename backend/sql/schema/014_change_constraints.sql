-- +goose Up
ALTER TABLE users
ADD CONSTRAINT name_email UNIQUE(name, email);

-- +goose Down
ALTER TABLE users
DROP CONSTRAINT name_email;
