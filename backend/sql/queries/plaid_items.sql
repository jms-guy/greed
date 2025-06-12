-- name: CreateItem :one
INSERT INTO plaid_items(id, user_id, item_id, access_token, request_id, created_at, updated_at)
VALUES (
    $1,
    $2,
    $3,
    ??,
    $5,
    NOW(),
    NOW()
)
RETURNING *;