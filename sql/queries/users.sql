-- name: CreateUser :one
INSERT INTO users(id, name, created_at, updated_at)
VALUES (
    $1,
    $2,
    NOW(),
    NOW()
)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;