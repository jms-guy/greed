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

-- name: DeleteUser :exec
DELETE FROM users
WHERE name = ?;

-- name: ClearUsers :exec
DELETE FROM users;