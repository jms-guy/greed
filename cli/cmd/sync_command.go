package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/models"
	"github.com/spf13/cobra"
)

// Syncs database with updated account balances, and transaction records
// Updates account balances, and deletes local transaction records before replacing them with
// updated records from server database.
func (app *CLIApp) commandSync(cmd *cobra.Command, args []string) error {
	itemName := args[0]
	itemsURL := app.Config.Client.BaseURL + "/api/items"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting credentials")
		return nil
	}

	// Get item ID and institution
	itemID, itemInst, err := findItemHelper(app, itemName, itemsURL)
	if err != nil {
		if strings.Contains(err.Error(), "no item found") {
			LogError(app.Config.Db, cmd, err, "No item found")
			return nil
		} else {
			LogError(app.Config.Db, cmd, err, "Error contacting server")
			return nil
		}
	}

	updateBalanceURL := app.Config.Client.BaseURL + "/api/items/" + itemID + "/access/balances"

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("PUT", updateBalanceURL, token, nil)
	})
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return nil
	}
	defer resp.Body.Close()

	var accUpdates models.Accounts

	serverErr := parseAndReturnServerError(resp)
	if serverErr != nil {
		LogError(app.Config.Db, cmd, serverErr, "Error contacting server")
		return nil
	}
	if err = json.NewDecoder(resp.Body).Decode(&accUpdates); err != nil {
		if err == io.EOF {
			LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
			return nil
		}
		LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
		return nil
	}

	fmt.Println(" > Syncing account balances...")

	for _, acc := range accUpdates.Accounts {
		avBalance := sql.NullFloat64{}
		if acc.AvailableBalance != "" {
			avBal, err := strconv.ParseFloat(acc.AvailableBalance, 64)
			if err != nil {
				LogError(app.Config.Db, cmd, fmt.Errorf("error converting string value: %w", err), "Data error")
				return nil
			}
			avBalance.Float64 = avBal
			avBalance.Valid = true
		}

		curBalance := sql.NullFloat64{}
		if acc.CurrentBalance != "" {
			curBal, err := strconv.ParseFloat(acc.CurrentBalance, 64)
			if err != nil {
				LogError(app.Config.Db, cmd, fmt.Errorf("error converting string value: %w", err), "Data error")
				return nil
			}
			curBalance.Float64 = curBal
			curBalance.Valid = true
		}
		params := database.UpsertAccountParams{
			ID:               acc.Id,
			CreatedAt:        time.Now().Format("2006-01-02"),
			UpdatedAt:        time.Now().Format("2006-01-02"),
			Name:             acc.Name,
			Type:             acc.Type,
			Subtype:          sql.NullString{String: acc.Subtype, Valid: true},
			Mask:             sql.NullString{String: acc.Mask, Valid: true},
			OfficialName:     sql.NullString{String: acc.OfficialName, Valid: true},
			AvailableBalance: avBalance,
			CurrentBalance:   curBalance,
			IsoCurrencyCode:  sql.NullString{String: acc.IsoCurrencyCode, Valid: true},
			InstitutionName:  sql.NullString{String: itemInst, Valid: true},
			UserID:           creds.User.ID.String(),
		}
		_, err = app.Config.Db.UpsertAccount(context.Background(), params)
		if err != nil {
			LogError(app.Config.Db, cmd, fmt.Errorf("error updating account record: %w", err), "Local database error")
			continue
		}
	}

	fmt.Println(" > Account balances synced successfully.")
	fmt.Println(" > Fetching transaction data...")

	err = processWebhookRecords(app, itemID, "ITEM", "DEFAULT_UPDATE")
	if err != nil {
		return err
	}

	syncTxnsURL := app.Config.Client.BaseURL + "/api/items/" + itemID + "/access/transactions"

	response, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("POST", syncTxnsURL, token, nil)
	})
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return nil
	}
	defer response.Body.Close()

	serverErr = parseAndReturnServerError(response)
	if serverErr != nil {
		LogError(app.Config.Db, cmd, serverErr, "Error contacting server")
		return nil
	}

	var txns []models.Transaction
	if err = json.NewDecoder(response.Body).Decode(&txns); err != nil {
		if err == io.EOF {
			LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
			return nil
		}
		LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
		return nil
	}

	fmt.Println(" > Syncing transaction records...")

	err = app.Config.Db.DeleteTransactions(context.Background(), creds.User.ID.String())
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error clearing local records: %w", err), "Local database error")
		return nil
	}

	for _, t := range txns {
		a, err := strconv.ParseFloat(t.Amount, 64)
		if err != nil {
			LogError(app.Config.Db, cmd, fmt.Errorf("error converting string value: %w", err), "Data error")
			return nil
		}

		params := database.CreateTransactionParams{
			ID:                      t.Id,
			AccountID:               t.AccountId,
			Amount:                  a,
			IsoCurrencyCode:         sql.NullString{String: t.IsoCurrencyCode, Valid: true},
			Date:                    sql.NullString{String: t.Date.Format("2006-01-02"), Valid: true},
			MerchantName:            sql.NullString{String: t.MerchantName, Valid: true},
			PaymentChannel:          t.PaymentChannel,
			PersonalFinanceCategory: t.PersonalFinanceCategory,
		}

		_, err = app.Config.Db.CreateTransaction(context.Background(), params)
		if err != nil {
			LogError(app.Config.Db, cmd, fmt.Errorf("error creating local records: %w", err), "Local database error")
			return nil
		}

		fmt.Printf("\r > %v", t.Id)
	}

	fmt.Println("\r > Transaction records synced successfully.")

	webhookCodes := []string{"TRANSACTIONS_UPDATES_AVAILABLE", "TRANSACTIONS_REMOVED", "DEFAULT_UPDATE", "INITIAL_UPDATE", "HISTORICAL_UPDATE", "SYNC_UPDATES_AVAILABLE"}

	for _, code := range webhookCodes {
		err = processWebhookRecords(app, itemID, code, "TRANSACTIONS")
		if err != nil {
			LogError(app.Config.Db, cmd, err, "Error contacting server")
			return nil
		}
	}

	return nil
}
