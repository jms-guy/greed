-- name: CreateAccount :one
INSERT INTO accounts(id, created_at, updated_at, balance, goal, currency, user_id)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: ResetAccounts :exec
DELETE FROM accounts;