-- name: CreateErrorLog :exec
INSERT INTO error_logs (timestamp, command, error_message)
VALUES (?, ?, ?)
RETURNING *;