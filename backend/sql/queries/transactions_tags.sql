-- name: CreateTransactionToTagRecord :exec
INSERT INTO transactions_to_tags (
    transaction_id,
    tag_id
)
VALUES (
    $1,
    $2
);

-- name: GetTransactionsWithTag :many
SELECT * FROM transactions_to_tags
WHERE tag_id = $1;

-- name: GetTagsForTransaction :many
SELECT * FROM transactions_to_tags
WHERE transaction_id = $1;