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
    institution_name)
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
    ?
)
RETURNING *; 

-- name: DeleteAccounts :exec
DELETE FROM accounts
WHERE institution_name = ?
AND user_id = ?;