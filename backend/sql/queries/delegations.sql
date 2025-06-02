-- name: CreateDelegation :one
INSERT INTO delegations (
    id,
    user_id,
    created_at,
    expires_at,
    revoked_at,
    is_revoked,
    last_used)
VALUES (
    $1,
    $2,
    NOW(),
    $3,
    NULL,
    FALSE,
    NOW()
)
RETURNING *;

-- name: GetDelegation :one
SELECT * FROM delegations
WHERE id = $1;

-- name: UpdateLastUsed :exec 
UPDATE delegations
SET last_used = NOW()
WHERE id = $1;

-- name: RevokeDelegationByID :exec
UPDATE delegations
SET revoked_at = NOW(), is_revoked = TRUE
WHERE id = $1;

-- name: RevokeDelegationByUser :exec
UPDATE delegations
SET revoked_at = NOW(), is_revoked = TRUE
WHERE user_id = $1;