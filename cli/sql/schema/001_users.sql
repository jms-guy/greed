-- +goose Up
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    hashed_password TEXT NOT NULL,
    email TEXT NOT NULL,
    is_verified BOOLEAN 
);

-- +goose Down
DROP TABLE users;