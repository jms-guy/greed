package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/plaid/plaid-go/v36/plaid"
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


//Db transaction for updating item transaction history, handling new, modified, and deleted data, 
//as well as updating the item's transaction cursor.
func ApplyTransactionUpdates(app *AppServer, ctx context.Context, added, modified []plaid.Transaction, removed []plaid.RemovedTransaction, cursor, itemID string) error {
	tx, err := app.Database.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error beginning database transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := app.Db.WithTx(tx)

	var (
		valueStrings 	[]string
		valueArgs 		[]interface{}
	)

	added = append(added, modified...)

	//This loop handles upserting transaction data
	for i, txn := range added {
		curCode := ""
		if txn.IsoCurrencyCode.IsSet(){
			curCode = *txn.IsoCurrencyCode.Get()
		}
		merchant := ""
		if txn.MerchantName.IsSet() {
			merchant = *txn.MerchantName.Get()
		}
		txnDate, _ := time.Parse("2006-01-02", txn.Date)

		n := i*8
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			n+1, n+2, n+3, n+4, n+5, n+6, n+7, n+8))
		valueArgs = append(valueArgs,
			txn.TransactionId, txn.AccountId, txn.Amount, curCode, txnDate, merchant, txn.PaymentChannel, txn.PersonalFinanceCategory,)
	}

	insertStmt := fmt.Sprintf(`
		INSERT INTO transactions (
			id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category 
		) VALUES %s
		ON CONFLICT (id) DO UPDATE SET
			account_id = EXCLUDED.account_id,
			amount = EXCLUDED.amount,
			iso_currency_code = EXCLUDED.iso_currency_code,
			date = EXCLUDED.date,
			merchant_name = EXCLUDED.merchant_name,
			payment_channel = EXCLUDED.payment_channel,
			personal_finance_category = EXCLUDED.personal_finance_category,
			updated_at = NOW()
	`, strings.Join(valueStrings, ","))

	_, err = tx.ExecContext(ctx, insertStmt, valueArgs...)
	if err != nil {
		return fmt.Errorf("error updating transaction records: %w", err)
	}

	if len(removed) > 0 {
		for _, txn := range removed {
			params := database.DeleteTransactionParams{
				ID: txn.TransactionId,
				AccountID: txn.AccountId,
			}

			err = qtx.DeleteTransaction(ctx, params)
			if err != nil {
				return fmt.Errorf("error deleting transaction record: %w", err)
			}
		}
	}
	updatedCursor := sql.NullString{
		String: cursor,
		Valid: true,
	}

	updateParams := database.UpdateCursorParams{
		TransactionSyncCursor: updatedCursor,
		ID: itemID,
	}
	err = qtx.UpdateCursor(ctx, updateParams)
	if err != nil {
		return fmt.Errorf("error updating account's transaction cursor: %w", err)
	}

	return tx.Commit()
}