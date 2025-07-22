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