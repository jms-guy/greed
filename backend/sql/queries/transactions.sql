-- name: CreateTransaction :one
INSERT INTO transactions (
    id, 
    created_at, 
    updated_at, 
    amount, 
    category, 
    description, 
    transaction_date, 
    transaction_type,
    currency_code,
    account_id)
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
    $8
)
RETURNING *;

-- name: UpdateTransactionDescription :exec
UPDATE transactions
SET description = $1, updated_at = now()
WHERE id = $2;

-- name: UpdateTransactionCategory :exec
UPDATE transactions
SET category = $1, updated_at = now()
WHERE id = $2;