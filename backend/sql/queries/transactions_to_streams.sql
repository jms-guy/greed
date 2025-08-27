-- name: CreateTransactionToStreamRecord :exec
INSERT INTO transactions_to_streams (
    transaction_id,
    stream_id
)
VALUES (
    $1,
    $2
);

-- name: GetTransactionsToStreamConnections :many
SELECT * FROM transactions_to_streams
WHERE stream_id = $1;