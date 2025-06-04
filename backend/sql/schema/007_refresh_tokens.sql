-- +goose Up 
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    hashed_token TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id)
    ON DELETE CASCADE,
    delegation_id UUID NOT NULL REFERENCES delegations(id)
    ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE refresh_tokens;