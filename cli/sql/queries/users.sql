-- name: CreateUser :one
INSERT INTO users(id, name, created_at, updated_at, hashed_password, email, is_verified)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE name = ?;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ?;

-- name: DeleteUser :exec
DELETE FROM users
WHERE name = ?;

-- name: ClearUsers :exec
DELETE FROM users;

-- name: VerifyEmail :exec
UPDATE users
SET is_verified = ?, updated_at = ?
WHERE name = ?;

-- name: UpdatePassword :exec
UPDATE users
SET hashed_password = ?, updated_at = ?
WHERE name = ?;