-- name: GetIncomeForMonth :one
SELECT CAST(
    CASE 
        WHEN COUNT(*) = 0 THEN 0
        ELSE SUM(amount) 
    END as NUMERIC(16,2)
) as monthly_income
FROM transactions
WHERE transaction_type = 'credit'
  AND transaction_date >= make_date($1, $2, 1)
  AND transaction_date < make_date($1, $2, 1) + interval '1 month'
  AND account_id = $3;

-- name: GetExpensesForMonth :one
SELECT CAST(
    CASE 
        WHEN COUNT(*) = 0 THEN 0
        ELSE SUM(amount) 
    END as NUMERIC(16,2)
) as monthly_expenses
FROM transactions
WHERE transaction_type = 'debit'
  AND transaction_date >= make_date($1, $2, 1)
  AND transaction_date < make_date($1, $2, 1) + interval '1 month'
  AND account_id = $3;

-- name: GetNetIncomeForMonth :one
SELECT CAST(
    CASE 
        WHEN COUNT(*) = 0 THEN 0
        ELSE SUM(amount) 
    END as NUMERIC(16,2)
) as net_income
FROM transactions
WHERE transaction_type != 'transfer'
  AND transaction_date >= make_date($1, $2, 1)
  AND transaction_date < make_date($1, $2, 1) + interval '1 month'
  AND account_id = $3;