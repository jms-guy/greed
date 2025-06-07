package handlers

import (
	"context"
	"fmt"
	"github.com/jms-guy/greed/backend/internal/database"
)

//Transaction for expiring a delegation session, if refresh token given is expired
func ExpireDelegation(app *AppServer, ctx context.Context, tokenHash string, token database.RefreshToken) error {
	tx, err := app.Database.Begin()
		if err != nil {
			return fmt.Errorf("error beginning database transaction: %w", err)
		}
		defer tx.Rollback()

		qtx := app.Db.WithTx(tx)

		err  = qtx.ExpireToken(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("error expiring refresh token: %w", err)
		}
		err = qtx.RevokeDelegationByID(ctx, token.DelegationID)
		if err != nil {
			return fmt.Errorf("error revoking session delegation: %w", err)
		}
		return tx.Commit()
}

//Transaction for revoking a delegation session, if a token was used more than once
func RevokeDelegation(app *AppServer, ctx context.Context, token database.RefreshToken) error {
	tx, err := app.Database.Begin()
	if err != nil {
		return fmt.Errorf("error beginning database transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := app.Db.WithTx(tx)

	err = qtx.RevokeDelegationByID(ctx, token.DelegationID)
	if err != nil {
		return fmt.Errorf("error revoking session delegation: %w", err)
	}
	err = qtx.ExpireAllDelegationTokens(ctx, token.DelegationID)
	if err != nil {
		return fmt.Errorf("error expiring tokens of delegation: %w", err)
	}
	return tx.Commit()
}

