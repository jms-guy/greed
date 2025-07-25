-- name: CreatePlaidWebhookRecord :one
INSERT INTO plaid_webhook_records(id, webhook_type, webhook_code, item_id, user_id, created_at, processed, processed_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    NOW(),
    FALSE,
    NULL
)
RETURNING *;

-- name: GetWebhookRecords :many
SELECT plaid_webhook_records.* FROM plaid_webhook_records
WHERE plaid_webhook_records.user_id = $1 AND plaid_webhook_records.processed = FALSE;


-- name: ProcessWebhookRecordsByType :exec
UPDATE plaid_webhook_records
SET processed = TRUE, processed_at = NOW()
WHERE item_id = $1
    AND user_id = $2
    AND webhook_type = $3
    AND webhook_code = $4
    AND created_at <= $5 
    AND processed = FALSE;
