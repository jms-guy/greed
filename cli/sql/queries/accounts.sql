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
    institution_name,
    user_id)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *; 

-- name: GetAccount :one
SELECT * FROM accounts
WHERE name = ?
AND user_id = ?;

-- name: GetAllAccounts :many
SELECT * FROM accounts
WHERE user_id = ?;

-- name: DeleteAccounts :exec
DELETE FROM accounts
WHERE institution_name = ?
AND user_id = ?;

-- name: UpdateAcc :exec 
UPDATE accounts
SET available_balance = ?, current_balance = ?
WHERE id = ? AND user_id = ?;