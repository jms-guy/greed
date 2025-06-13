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
    plaid_account_id, 
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
    $10,
    $11
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


