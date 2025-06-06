-- name: CreateVerificationRecord :one
INSERT INTO verification_records (user_id, verification_code, expiry_time)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetVerificationRecord :one
SELECT * FROM verification_records
WHERE verification_code = $1;


-- name: DeleteVerificationRecord :exec
DELETE FROM verification_records
WHERE verification_code = $1;