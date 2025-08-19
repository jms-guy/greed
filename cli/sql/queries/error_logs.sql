-- name: CreateErrorLog :exec
INSERT INTO error_logs (timestamp, command, error_message)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetLogs :many
SELECT * FROM error_logs
ORDER BY timestamp DESC
LIMIT 5;