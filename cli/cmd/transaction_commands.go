package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"slices"
	"strconv"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/cli/internal/tables"
	"github.com/jms-guy/greed/cli/internal/utils"
	"github.com/jms-guy/greed/models"
	"github.com/spf13/cobra"
)

// Get transaction records for given account from the server database.
// Takes into account optional flags, creating a dynamic query to retrieve and sort the data on.
// If summary flag is present, overrides most other flags, and returns a transaction summary instead
func (app *CLIApp) commandGetTxnsAccount(cmd *cobra.Command, args []string, merchant, category, channel, date, start, end, order string, min, max, limit, pageSize int, summary bool) error {
	var err error
	queryString := utils.BuildQueries(merchant, category, channel, date, start, end, min, max, limit, summary)

	var account database.Account

	// Get account to use in command
	if len(args) == 0 && app.Config.Settings.DefaultAccount.ID == "" {
		LogError(app.Config.Db, cmd, fmt.Errorf("no account given"), "Missing argument")
		return nil

	} else if len(args) == 1 {
		var err error
		accountName := args[0]

		creds, err := auth.GetCreds(app.Config.ConfigFP)
		if err != nil {
			LogError(app.Config.Db, cmd, err, "Error getting credentials")
			return err
		}

		params := database.GetAccountParams{
			Name:   accountName,
			UserID: creds.User.ID.String(),
		}
		account, err = app.Config.Db.GetAccount(context.Background(), params)
		if err != nil {
			LogError(app.Config.Db, cmd, fmt.Errorf("error getting local account: %w", err), "Local database error")
			return err
		}

	} else {
		account = app.Config.Settings.DefaultAccount
	}

	// Get queried transactions from server
	txnsURL := app.Config.Client.BaseURL + "/api/accounts/" + account.ID + "/transactions"
	if queryString != "?" {
		txnsURL = txnsURL + queryString
	}

	res, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", txnsURL, token, nil)
	})
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http request: %w", err), "Error contacting server")
		return err
	}
	defer res.Body.Close()

	err = checkResponseStatus(res)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	// If summary flag is present, decode into summary struct and print table
	if summary {
		var summaries []models.MerchantSummary
		if err = json.NewDecoder(res.Body).Decode(&summaries); err != nil {
			LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
			return err
		}

		err = tables.PaginateSummariesTable(summaries, account.Name, merchant, pageSize)
		if err != nil {
			LogError(app.Config.Db, cmd, fmt.Errorf("error creating transactions table: %w", err), "Error drawing table")
			return err
		}

		return nil
	}

	// Else decode into regular transaction struct, calculate balance and determine filters
	var txns []models.Transaction
	if err = json.NewDecoder(res.Body).Decode(&txns); err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
		return err
	}

	currentBalance := account.CurrentBalance.Float64
	var historicalBalances []float64

	runningBalance := currentBalance

	// Get recurring transaction data
	recurring, err := app.GetRecurringData(account.ID)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting transaction data")
		return err
	}

	for _, txn := range txns {

		amountFloat, _ := strconv.ParseFloat(txn.Amount, 64)
		historicalBalances = append(historicalBalances, runningBalance)

		runningBalance += amountFloat
	}

	if order == "ASC" {
		slices.Reverse(txns)
		slices.Reverse(historicalBalances)
	}

	isFiltered := true
	if merchant == "" && category == "" && channel == "" && date == "" && start == "" && end == "" && min == math.MinInt64 && max == math.MaxInt64 {
		isFiltered = false
	}

	// Draw paginated transactions table
	err = tables.PaginateTransactionsTable(txns, account.Name, historicalBalances, pageSize, isFiltered, recurring)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error creating transactions table: %w", err), "Error drawing table")
		return err
	}

	return nil
}

// Function gets recurring transaction data from server for account
func (app *CLIApp) GetRecurringData(accountID string) (models.RecurringData, error) {
	var recurringData models.RecurringData

	recurringURL := app.Config.Client.BaseURL + "/api/accounts/" + accountID + "/transactions/recurring"

	res, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", recurringURL, token, nil)
	})
	if err != nil {
		return recurringData, fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	err = checkResponseStatus(res)
	if err != nil {
		return recurringData, err
	}

	if err = json.NewDecoder(res.Body).Decode(&recurringData); err != nil {
		return recurringData, fmt.Errorf("decoding err: %w", err)
	}

	return recurringData, nil
}
