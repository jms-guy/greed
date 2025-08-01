-- name: CreateTag :one
INSERT INTO transaction_tags (
    id,
    name,
    user_id,
    created_at
)
VALUES (
    $1,
    $2,
    $3,
    NOW()
)
RETURNING *;

-- name: GetTag :one
SELECT * FROM transaction_tags
WHERE name = $1
AND user_id = $2;

-- name: GetAllTagsForUser :many
SELECT * FROM transaction_tags
WHERE user_id = $1;

-- name: DeleteTag :exec
DELETE FROM transaction_tags
WHERE name = $1
AND user_id = $2;

-- name: DeleteAllTagsForUser :exec
DELETE FROM transaction_tags
WHERE user_id = $1;