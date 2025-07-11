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

-- name: VerifyUser :exec
UPDATE users
SET is_verified = true, updated_at = now()
WHERE id = $1;

-- name: UpdatePassword :exec
UPDATE users
SET hashed_password = $1, updated_at = now()
WHERE id = $2;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateMember :exec
UPDATE users
SET is_member = true, updated_at = NOW()
WHERE id = $1;

-- name: UpdateFreeCalls :exec
UPDATE users
SET free_calls = free_calls - 1, updated_at = NOW()
WHERE id = $1;