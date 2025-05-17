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