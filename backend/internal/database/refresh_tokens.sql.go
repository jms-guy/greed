// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: refresh_tokens.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createToken = `-- name: CreateToken :one
INSERT INTO refresh_tokens (
    id,
    hashed_token,
    user_id,
    delegation_id,
    created_at,
    expires_at,
    is_used,
    used_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    NOW(),
    $5,
    FALSE,
    NULL
)
RETURNING id, hashed_token, user_id, delegation_id, created_at, expires_at, is_used, used_at
`

type CreateTokenParams struct {
	ID           uuid.UUID
	HashedToken  string
	UserID       uuid.UUID
	DelegationID uuid.UUID
	ExpiresAt    time.Time
}

func (q *Queries) CreateToken(ctx context.Context, arg CreateTokenParams) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, createToken,
		arg.ID,
		arg.HashedToken,
		arg.UserID,
		arg.DelegationID,
		arg.ExpiresAt,
	)
	var i RefreshToken
	err := row.Scan(
		&i.ID,
		&i.HashedToken,
		&i.UserID,
		&i.DelegationID,
		&i.CreatedAt,
		&i.ExpiresAt,
		&i.IsUsed,
		&i.UsedAt,
	)
	return i, err
}

const expireAllDelegationTokens = `-- name: ExpireAllDelegationTokens :exec
UPDATE refresh_tokens
SET is_used = TRUE, used_at = NOW()
WHERE delegation_id = $1 AND is_used = FALSE
`

func (q *Queries) ExpireAllDelegationTokens(ctx context.Context, delegationID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, expireAllDelegationTokens, delegationID)
	return err
}

const expireToken = `-- name: ExpireToken :exec
UPDATE refresh_tokens
SET is_used = TRUE, used_at = NOW()
WHERE hashed_token = $1
`

func (q *Queries) ExpireToken(ctx context.Context, hashedToken string) error {
	_, err := q.db.ExecContext(ctx, expireToken, hashedToken)
	return err
}

const getToken = `-- name: GetToken :one
SELECT id, hashed_token, user_id, delegation_id, created_at, expires_at, is_used, used_at FROM refresh_tokens
WHERE hashed_token = $1
`

func (q *Queries) GetToken(ctx context.Context, hashedToken string) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, getToken, hashedToken)
	var i RefreshToken
	err := row.Scan(
		&i.ID,
		&i.HashedToken,
		&i.UserID,
		&i.DelegationID,
		&i.CreatedAt,
		&i.ExpiresAt,
		&i.IsUsed,
		&i.UsedAt,
	)
	return i, err
}
