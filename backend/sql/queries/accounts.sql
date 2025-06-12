-- name: CreateAccount :one
INSERT INTO accounts(id, created_at, updated_at, name, type, mask, official_name, plaid_account_id, item_id)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;

-- name: ResetAccounts :exec
DELETE FROM accounts;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE name = $1;

-- name: GetAllAccountsForUser :many
SELECT * FROM accounts
WHERE item_id = $1;


