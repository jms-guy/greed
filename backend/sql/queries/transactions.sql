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

-- name: ClearTransactionsTable :exec
DELETE FROM transactions;

-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE id = $1 AND account_id = $2;

-- name: DeleteTransactionsForAccount :exec
DELETE FROM transactions
WHERE account_id = $1;


-- name: GetTransactionsForUser :many
SELECT t.* 
FROM transactions AS t
INNER JOIN accounts AS a ON t.account_id = a.id
WHERE a.user_id = $1;