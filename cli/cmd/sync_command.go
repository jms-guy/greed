package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/models"
)

// Syncs database with updated account balances, and transaction records
// Updates account balances, and deletes local transaction records before replacing them with
// updated records from server database.
func (app *CLIApp) commandSync(args []string) error {
	itemName := args[0]
	itemsURL := app.Config.Client.BaseURL + "/api/items"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	// Get item ID and institution
	itemID, itemInst, err := findItemHelper(app, itemName, itemsURL)
	if err != nil {
		return err
	}

	updateBalanceURL := app.Config.Client.BaseURL + "/api/items/" + itemID + "/access/balances"

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("PUT", updateBalanceURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer resp.Body.Close()

	var accUpdates models.Accounts

	serverErr := parseAndReturnServerError(resp)
	if serverErr != nil {
		return serverErr
	}
	if err = json.NewDecoder(resp.Body).Decode(&accUpdates); err != nil {
		if err == io.EOF {
			return fmt.Errorf("decoding error: received an empty or incomplete response for accounts. This often indicates a temporary issue with Plaid or the financial institution. Please try syncing again later")
		}
		return fmt.Errorf("decoding error for account balances: %w. The response might be malformed or unexpected. Please try syncing again later", err)
	}

	fmt.Println(" > Syncing account balances...")

	for _, acc := range accUpdates.Accounts {
		avBalance := sql.NullFloat64{}
		if acc.AvailableBalance != "" {
			avBal, err := strconv.ParseFloat(acc.AvailableBalance, 64)
			if err != nil {
				return fmt.Errorf("error converting balance to float: %w", err)
			}
			avBalance.Float64 = avBal
			avBalance.Valid = true
		}

		curBalance := sql.NullFloat64{}
		if acc.CurrentBalance != "" {
			curBal, err := strconv.ParseFloat(acc.CurrentBalance, 64)
			if err != nil {
				return fmt.Errorf("error converting balance to float: %w", err)
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
			fmt.Printf("Error updating balance of acc #%v: %s\n", acc.Id, err)
			continue
		}
	}

	fmt.Println(" > Account balances synced successfully.")

	err = processWebhookRecords(app, itemID, "ITEM", "DEFAULT_UPDATE")
	if err != nil {
		return err
	}

	syncTxnsURL := app.Config.Client.BaseURL + "/api/items/" + itemID + "/access/transactions"

	response, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("POST", syncTxnsURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer response.Body.Close()

	serverErr = parseAndReturnServerError(response)
	if serverErr != nil {
		return serverErr
	}

	var txns []models.Transaction
	if err = json.NewDecoder(response.Body).Decode(&txns); err != nil {
		if err == io.EOF {
			return fmt.Errorf("decoding error: received an empty or incomplete response for transactions. This often indicates a temporary issue with Plaid or the financial institution. Please try syncing again later")
		}
		return fmt.Errorf("decoding error for transactions: %w. The response might be malformed or unexpected. Please try syncing again later", err)
	}

	fmt.Println(" > Syncing transaction records...")

	err = app.Config.Db.DeleteTransactions(context.Background(), creds.User.ID.String())
	if err != nil {
		return fmt.Errorf("error clearing local records: %w", err)
	}

	for _, t := range txns {
		a, err := strconv.ParseFloat(t.Amount, 64)
		if err != nil {
			return fmt.Errorf("error converting string value: %w", err)
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
			return fmt.Errorf("error creating local record: %w", err)
		}

		fmt.Printf("\r > %v", t.Id)
	}

	fmt.Println("\r > Transaction records synced successfully.")

	err = processWebhookRecords(app, itemID, "TRANSACTIONS", "TRANSACTIONS_UPDATES_AVAILABLE")
	if err != nil {
		return err
	}
	err = processWebhookRecords(app, itemID, "TRANSACTIONS", "TRANSACTIONS_REMOVED")
	if err != nil {
		return err
	}
	err = processWebhookRecords(app, itemID, "TRANSACTIONS", "DEFAULT_UPDATE")
	if err != nil {
		return err
	}

	return nil
}
