package main
/*
import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
)

type Database interface {
	GetToken(ctx context.Context, hashedToken string) (database.RefreshToken, error)
	ExpireToken(ctx context.Context, hashedToken string) error
	RevokeDelegationByID(ctx context.Context, id uuid.UUID) error
	ExpireAllDelegationTokens(ctx context.Context, id uuid.UUID) error

	BeginTx(ctx context.Context) (Tx, error)
}

type Tx interface {
	Commit() error
	Rollback() error

	GetToken(ctx context.Context, hashedToken string) (database.RefreshToken, error)
	ExpireToken(ctx context.Context, hashedToken string) error
	RevokeDelegationByID(ctx context.Context, id uuid.UUID) error
	ExpireAllDelegationTokens(ctx context.Context, id uuid.UUID) error
}

type DB struct {
	db *sql.DB
}

type DBTx struct {
	tx *sql.Tx
}

func (d *DB) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &DBTx{tx: tx}, nil
}

func (t *DBTx) Commit() error {
	return t.tx.Commit()
}

func (t *DBTx) Rollback() error {
	return t.tx.Rollback()
}*/