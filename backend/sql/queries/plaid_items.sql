-- name: CreateItem :one
INSERT INTO plaid_items(id, user_id, access_token, institution_name, nickname, transaction_sync_cursor, created_at, updated_at)
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
WHERE id = $1;

-- name: GetItemByName :one
SELECT * FROM plaid_items
WHERE nickname = $1;

-- name: GetItemsByUser :many
SELECT * FROM plaid_items
WHERE user_id = $1;

-- name: GetItemByID :one
SELECT * FROM plaid_items
WHERE id = $1;

-- name: UpdateNickname :exec
UPDATE plaid_items
SET nickname = $1, updated_at = NOW()
WHERE id = $2 AND user_id = $3;

-- name: DeleteItem :exec
DELETE FROM plaid_items
WHERE id = $1 AND user_id = $2;

-- name: ResetItems :exec
DELETE FROM plaid_items;

-- name: GetLatestCursorOrNil :one
SELECT transaction_sync_cursor FROM plaid_items
WHERE id = $1;

-- name: UpdateCursor :exec
UPDATE plaid_items
SET transaction_sync_cursor = $1, updated_at = NOW()
WHERE id = $2;

-- name: GetCursor :one
SELECT transaction_sync_cursor FROM plaid_items
WHERE id = $1 AND access_token = $2;