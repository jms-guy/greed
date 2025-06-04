-- name: CreateUser :one
INSERT INTO users(id, name, created_at, updated_at, hashed_password, email, is_verified)
VALUES (
    $1,
    $2,
    NOW(),
    NOW(),
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1;

-- name: GetAllUsers :many
SELECT name FROM users;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: ResetUsers :exec
DELETE FROM users;