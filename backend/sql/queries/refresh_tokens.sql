-- name: CreateToken :one
INSERT INTO refresh_tokens (
    id,
    hashed_token,
    user_id,
    delegation_id,
    created_at,
    expires_at,
    is_used,
    used_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    NOW(),
    $5,
    FALSE,
    NULL
)
RETURNING *;

-- name: GetToken :one
SELECT * FROM refresh_tokens
WHERE hashed_token = $1;

-- name: ExpireToken :exec
UPDATE refresh_tokens
SET is_used = TRUE, used_at = NOW()
WHERE hashed_token = $1;

-- name: ExpireAllDelegationTokens :exec 
UPDATE refresh_tokens
SET is_used = TRUE, used_at = NOW()
WHERE delegation_id = $1 AND is_used = FALSE;