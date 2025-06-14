-- name: CreateAccount :one
INSERT INTO accounts(
    id,
    created_at, 
    updated_at, 
    name, 
    type, 
    subtype, 
    mask, 
    official_name,
    available_balance,
    current_balance,
    iso_currency_code, 
    item_id)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
)
RETURNING *;

-- name: ResetAccounts :exec
DELETE FROM accounts;

-- name: GetAccountById :one
SELECT * FROM accounts
WHERE id = $1 AND user_id = $2;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE name = $1;

-- name: GetAllAccountsForUser :many
SELECT * FROM accounts
WHERE user_id = $1;

-- name: GetAccountsForItem :many
SELECT * FROM accounts
WHERE item_id = $1;

-- name: UpdateBalances :one
UPDATE accounts
SET available_balance = $1, current_balance = $2, updated_at = NOW()
WHERE id = $3 AND item_id = $4
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1 AND user_id = $2;