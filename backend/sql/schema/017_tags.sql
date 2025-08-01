-- +goose Up 
CREATE TABLE transaction_tags (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id)
    ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT UC_Name_User UNIQUE (name, user_id)
);

-- +goose Down
DROP TABLE transaction_tags;