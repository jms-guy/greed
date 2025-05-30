-- name: CreateDelegation :one
INSERT INTO delegations (
    id,
    user_id,
    created_at,
    revoked_at,
    is_revoked,
    last_used)
VALUES (
    $1,
    $2,
    NOW(),
    NULL,
    FALSE,
    NOW()
)
RETURNING *;

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