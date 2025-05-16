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

-- name: GetAccount :one
SELECT * FROM accounts
WHERE user_id = $1;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE user_id = $1;

-- name: UpdateBalance :exec
UPDATE accounts
SET balance = $1, updated_at = now()
WHERE id = $2;

-- name: UpdateGoal :exec
UPDATE accounts
SET goal = $1, updated_at = now()
WHERE id = $2;

-- name: UpdateCurrency :exec
UPDATE accounts
SET currency = $1, updated_at = now()
WHERE id = $2;