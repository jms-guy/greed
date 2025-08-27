-- name: CreateStream :exec
INSERT INTO recurring_streams(
    id,
    account_id,
    description,
    merchant_name,
    frequency,
    is_active,
    predicted_next_date,
    stream_type,
    created_at
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    NOW()
);

-- name: GetStreamsForAcc :many
SELECT * FROM recurring_streams
WHERE account_id = $1;