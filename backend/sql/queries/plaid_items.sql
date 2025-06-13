-- name: CreateItem :one
INSERT INTO plaid_items(id, user_id, item_id, access_token, request_id, nickname, created_at, updated_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    NOW(),
    NOW()
)
RETURNING *;

-- name: GetAccessToken :one
SELECT * FROM plaid_items
WHERE item_id = $1;

-- name: GetItemByName :one
SELECT * FROM plaid_items
WHERE nickname = $1;

-- name: GetItemsByUser :many
SELECT * FROM plaid_items
WHERE user_id = $1;