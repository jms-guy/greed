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

-- name: GetIncomeForMonth :one
SELECT CAST(SUM(amount) as NUMERIC(16,2)) monthly_income
FROM transactions
WHERE
    transaction_type = 'credit'
    AND transaction_date >= make_date($1, $2, 1)
    AND transaction_date < make_date($1, $2, 1) + interval '1 month'
    AND account_id = $3;

-- name: GetExpensesForMonth :one
SELECT CAST(SUM(amount) as NUMERIC(16,2)) monthly_expenses
FROM transactions
WHERE
    transaction_type = 'debit'  
    AND transaction_date >= make_date($1, $2, 1)
    AND transaction_date < make_date($1, $2, 1) + interval '1 month'
    AND account_id = $3;

-- name: GetNetIncomeForMonth :one
SELECT CAST(SUM(amount) as NUMERIC(16,2)) net_income
FROM transactions
WHERE
    transaction_type != 'transfer'
    AND transaction_date >= make_date($1, $2, 1)
    AND transaction_date < make_date($1, $2, 1) + interval '1 month'
    AND account_id = $3;