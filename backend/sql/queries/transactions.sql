-- name: CreateTransaction :one
INSERT INTO transactions (
    id, 
    account_id,
    amount,
    iso_currency_code,
    date,
    merchant_name,
    payment_channel,
    personal_finance_category,
    created_at,
    updated_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    NOW(),
    NOW()
)
RETURNING *;

-- name: GetTransactions :many
SELECT * FROM transactions
WHERE account_id = $1;

-- name: GetTransactionsByMerchant :many
SELECT * FROM transactions
WHERE account_id = $1 AND merchant_name = $2;

-- name: GetTransactionsByChannel :many
SELECT * FROM transactions
WHERE account_id = $1 AND payment_channel = $2;

-- name: GetTransactionsByCategory :many
SELECT * FROM transactions
WHERE account_id = $1 AND personal_finance_category = $2;

-- name: GetTransactionsforDate :many
SELECT * FROM transactions
WHERE account_id = $1 AND date = $2;

-- name: ClearTransactionsTable :exec
DELETE FROM transactions;

-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE id = $1 AND account_id = $2;

-- name: DeleteTransactionsForAccount :exec
DELETE FROM transactions
WHERE account_id = $1;
