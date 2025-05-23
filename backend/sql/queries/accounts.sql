-- name: CreateAccount :one
INSERT INTO accounts(id, created_at, updated_at, name, input_type, currency, user_id)
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
WHERE id = $1
AND user_id = $2;

-- name: GetAllAccountsForUser :many
SELECT * FROM accounts
WHERE user_id = $1;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1
AND user_id = $2;

-- name: UpdateCurrency :one
UPDATE accounts
SET currency = $1, updated_at = now()
WHERE id = $2
RETURNING *;