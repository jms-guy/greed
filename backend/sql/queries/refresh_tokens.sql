-- name: CreateToken :one
INSERT INTO refresh_tokens (
    id,
    hashed_token,
    delegation_id,
    created_at,
    expires_at,
    is_used,
    used_at)
VALUES (
    $1,
    $2,
    $3,
    NOW(),
    $4,
    FALSE,
    NULL
)
RETURNING *;

-- name: ExpireToken :exec
UPDATE refresh_tokens
SET is_used = TRUE, used_at = NOW()
WHERE hashed_token = $1;