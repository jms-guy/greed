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