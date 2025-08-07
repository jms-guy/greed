-- +goose Up
ALTER TABLE users
DROP CONSTRAINT name_email;

ALTER TABLE users
ADD CONSTRAINT UC_email UNIQUE (email);

-- +goose Down
ALTER TABLE users
DROP CONSTRAINT UC_email;

ALTER TABLE users
ADD CONSTRAINT name_email UNIQUE(name, email);