// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: accounts.sql

package database

import (
	"context"
	"database/sql"
)

const createAccount = `-- name: CreateAccount :one
INSERT INTO accounts(
    id,
    created_at,
    updated_at,
    name, 
    type,
    subtype,
    mask, 
    official_name,
    available_balance, 
    current_balance, 
    iso_currency_code, 
    institution_name,
    user_id)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING id, created_at, updated_at, name, type, subtype, mask, official_name, available_balance, current_balance, iso_currency_code, institution_name, user_id
`

type CreateAccountParams struct {
	ID               string
	CreatedAt        string
	UpdatedAt        string
	Name             string
	Type             string
	Subtype          sql.NullString
	Mask             sql.NullString
	OfficialName     sql.NullString
	AvailableBalance sql.NullFloat64
	CurrentBalance   sql.NullFloat64
	IsoCurrencyCode  sql.NullString
	InstitutionName  sql.NullString
	UserID           string
}

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, createAccount,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Name,
		arg.Type,
		arg.Subtype,
		arg.Mask,
		arg.OfficialName,
		arg.AvailableBalance,
		arg.CurrentBalance,
		arg.IsoCurrencyCode,
		arg.InstitutionName,
		arg.UserID,
	)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Type,
		&i.Subtype,
		&i.Mask,
		&i.OfficialName,
		&i.AvailableBalance,
		&i.CurrentBalance,
		&i.IsoCurrencyCode,
		&i.InstitutionName,
		&i.UserID,
	)
	return i, err
}

const deleteAccounts = `-- name: DeleteAccounts :exec
DELETE FROM accounts
WHERE institution_name = ?
AND user_id = ?
`

type DeleteAccountsParams struct {
	InstitutionName sql.NullString
	UserID          string
}

func (q *Queries) DeleteAccounts(ctx context.Context, arg DeleteAccountsParams) error {
	_, err := q.db.ExecContext(ctx, deleteAccounts, arg.InstitutionName, arg.UserID)
	return err
}

const getAccount = `-- name: GetAccount :one
SELECT id, created_at, updated_at, name, type, subtype, mask, official_name, available_balance, current_balance, iso_currency_code, institution_name, user_id FROM accounts
WHERE name = ?
AND user_id = ?
`

type GetAccountParams struct {
	Name   string
	UserID string
}

func (q *Queries) GetAccount(ctx context.Context, arg GetAccountParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, getAccount, arg.Name, arg.UserID)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Type,
		&i.Subtype,
		&i.Mask,
		&i.OfficialName,
		&i.AvailableBalance,
		&i.CurrentBalance,
		&i.IsoCurrencyCode,
		&i.InstitutionName,
		&i.UserID,
	)
	return i, err
}

const getAllAccounts = `-- name: GetAllAccounts :many
SELECT id, created_at, updated_at, name, type, subtype, mask, official_name, available_balance, current_balance, iso_currency_code, institution_name, user_id FROM accounts
WHERE user_id = ?
`

func (q *Queries) GetAllAccounts(ctx context.Context, userID string) ([]Account, error) {
	rows, err := q.db.QueryContext(ctx, getAllAccounts, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Account
	for rows.Next() {
		var i Account
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Type,
			&i.Subtype,
			&i.Mask,
			&i.OfficialName,
			&i.AvailableBalance,
			&i.CurrentBalance,
			&i.IsoCurrencyCode,
			&i.InstitutionName,
			&i.UserID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateAcc = `-- name: UpdateAcc :exec
UPDATE accounts
SET available_balance = ?, current_balance = ?
WHERE id = ? AND user_id = ?
`

type UpdateAccParams struct {
	AvailableBalance sql.NullFloat64
	CurrentBalance   sql.NullFloat64
	ID               string
	UserID           string
}

func (q *Queries) UpdateAcc(ctx context.Context, arg UpdateAccParams) error {
	_, err := q.db.ExecContext(ctx, updateAcc,
		arg.AvailableBalance,
		arg.CurrentBalance,
		arg.ID,
		arg.UserID,
	)
	return err
}
