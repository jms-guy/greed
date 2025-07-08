-- name: CreateTransaction :one
INSERT INTO transactions(
    id, 
    account_id, 
    amount, 
    iso_currency_code, 
    date, 
    merchant_name, 
    payment_channel, 
    personal_finance_category
    )
VALUES (
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

-- name: GetTransactions :many
SELECT * FROM transactions
WHERE account_id = ?;

-- name: DeleteTransactions :exec
DELETE FROM transactions
WHERE account_id IN (
    SELECT id FROM accounts
    WHERE user_id = ?
);