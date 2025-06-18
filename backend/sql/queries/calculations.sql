-- name: GetDataForMonth :one
SELECT
    CAST(SUM(CASE WHEN amount < 0 THEN amount ELSE 0 END) AS NUMERIC(16, 2)) AS income,
    CAST(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END) AS NUMERIC(16, 2)) AS expenses,
    CAST(SUM(amount) AS NUMERIC(16, 2)) AS net_income
FROM transactions
WHERE date >= make_date($1, $2, 1)
  AND date < (make_date($1, $2, 1) + interval '1 month')
  AND account_id = $3;

-- name: GetDataForAllMonths :many
SELECT
  EXTRACT(YEAR FROM date)::int AS year,
  EXTRACT(MONTH FROM date)::int AS month,
  CAST(SUM(CASE WHEN amount < 0 THEN amount ELSE 0 END) AS NUMERIC(16,2)) AS income,
  CAST(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END) AS NUMERIC(16,2)) AS expenses,
  CAST(SUM(amount) AS NUMERIC(16,2)) AS net_income
FROM transactions
WHERE account_id = $1
GROUP BY year, month
ORDER BY year, month;