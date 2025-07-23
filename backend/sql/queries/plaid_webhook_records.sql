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

-- name: GetWebhookRecord :many
SELECT * FROM plaid_webhook_records
WHERE user_id = $1 AND item_id = $2;

-- name: DeleteWebhookRecord :exec
DELETE FROM plaid_webhook_records
WHERE user_id = $1 AND item_id = $2;