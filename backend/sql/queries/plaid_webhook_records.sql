-- name: CreatePlaidWebhookRecord :one
INSERT INTO plaid_webhook_records(id, webhook_type, webhook_code, item_id, user_id, created_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    NOW()
)
RETURNING *;

-- name: GetWebhookRecords :many
SELECT plaid_webhook_records.* FROM plaid_webhook_records
INNER JOIN users ON users.id = plaid_webhook_records.user_id
WHERE plaid_webhook_records.user_id = $1;


-- name: DeleteWebhookRecord :exec
DELETE FROM plaid_webhook_records
WHERE user_id = $1 AND item_id = $2;