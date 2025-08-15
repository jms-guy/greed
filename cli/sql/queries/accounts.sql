-- name: UpsertAccount :one
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
ON CONFLICT (id) DO UPDATE SET 
    available_balance = EXCLUDED.available_balance,
    current_balance = EXCLUDED.current_balance,
    updated_at = NOW()
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
