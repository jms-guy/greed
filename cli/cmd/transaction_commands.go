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
)

//Get transaction records for given account from the server database.
//Takes into account optional flags, creating a dynamic query to retrieve and sort the data on.
//If summary flag is present, overrides most other flags, and returns a transaction summary instead
func (app *CLIApp) commandGetTxnsAccount(accountName, merchant, category, channel, date, start, end, order string, min, max, limit, pageSize int, summary bool) error {

	var err error
	queryString := utils.BuildQueries(merchant, category, channel, date, start, end, min, max, limit, summary)

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	params := database.GetAccountParams{
		Name: accountName,
		UserID: creds.User.ID.String(),
	}
	account, err := app.Config.Db.GetAccount(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error getting local account record: %w", err)
	}

	txnsURL := app.Config.Client.BaseURL + "/api/accounts/" + account.ID + "/transactions"
	if queryString != "?" {
		txnsURL = txnsURL + queryString
	}

	res, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", txnsURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	checkResponseStatus(res)
	
	if summary {
		var summaries []models.MerchantSummary
		if err = json.NewDecoder(res.Body).Decode(&summaries); err != nil {
			return fmt.Errorf("decoding error: %w", err)
		}

		err = tables.PaginateSummariesTable(summaries, accountName, merchant, pageSize)
		if err != nil {
			return fmt.Errorf("error creating transactions table: %w", err)
		}

		return nil
	}

	var txns []models.Transaction
	if err = json.NewDecoder(res.Body).Decode(&txns); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	currentBalance := account.CurrentBalance.Float64
	var historicalBalances []float64

	runningBalance := currentBalance 

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

	err = tables.PaginateTransactionsTable(txns, accountName, historicalBalances, pageSize, isFiltered)
	if err != nil {
		return fmt.Errorf("error creating transactions table: %w", err)
	}

	return nil
}