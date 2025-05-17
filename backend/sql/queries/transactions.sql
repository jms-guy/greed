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

-- name: GetSingleTransaction :one
SELECT * FROM transactions
WHERE id = $1;

-- name: GetTransactions :many
SELECT * FROM transactions
WHERE account_id = $1;

-- name: GetTransactionsOfType :many
SELECT * FROM transactions
WHERE account_id = $1
AND transaction_type = $2;

-- name: GetTransactionsOfCategory :many
SELECT * FROM transactions
WHERE account_id = $1
AND category = $2;

-- name: ClearTransactionsTable :exec
DELETE FROM transactions;

-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE id = $1;

-- name: DeleteTransactionsForAccount :exec
DELETE FROM transactions
WHERE account_id = $1;

-- name: DeleteTransactionsOfType :exec
DELETE FROM transactions
WHERE account_id = $1
AND transaction_type = $2;

-- name: DeleteTransactionsOfCategory :exec
DELETE FROM transactions
WHERE account_id = $1
AND category = $2;