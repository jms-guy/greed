package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/cli/internal/tables"
	"github.com/jms-guy/greed/cli/internal/utils"
	"github.com/jms-guy/greed/models"
)

//Get transaction records for given account
//Currently basic, add query arguments, create local txn records
func (app *CLIApp) commandGetTxnsAccount(accountName, merchant, category, channel, date, start, end, order string, min, max, limit int, summary bool) error {

	var err error
	queryString := utils.BuildQueries(merchant, category, channel, date, start, end, order, min, max, limit, summary)

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	user, err := app.Config.Db.GetUser(context.Background(), creds.User.Name)
	if err != nil {
		return fmt.Errorf("error getting local user record: %w", err)
	}

	params := database.GetAccountParams{
		Name: accountName,
		UserID: user.ID,
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

		tbl := tables.MakeTableForSummaries(summaries, accountName)
		tbl.Print()

		return nil
	}

	var txns []models.Transaction
	if err = json.NewDecoder(res.Body).Decode(&txns); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	tbl := tables.MakeTableForTransactions(txns, accountName)
	tbl.Print()

	return nil
}