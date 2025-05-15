-- name: ValidateCurrency :one
SELECT EXISTS(SELECT 1 FROM supported_currencies WHERE code = $1 AND active = true);