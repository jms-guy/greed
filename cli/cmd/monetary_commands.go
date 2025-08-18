package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/charts"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/cli/internal/tables"
	"github.com/jms-guy/greed/models"
	"github.com/spf13/cobra"
)

// Gets an account's income/expense data through querying server database transaction data.
// Displays data in a visual format based on flag value passed through mode
func (app *CLIApp) commandGetIncome(cmd *cobra.Command, accountName, mode string) error {
	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting credentials")
		return nil
	}

	params := database.GetAccountParams{
		Name:   accountName,
		UserID: creds.User.ID.String(),
	}
	account, err := app.Config.Db.GetAccount(context.Background(), params)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error getting local account record: %w", err), "Local database error")
		return nil
	}

	incURL := app.Config.Client.BaseURL + "/api/accounts/" + account.ID + "/transactions/monetary"

	res, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", incURL, token, nil)
	})
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return nil
	}
	defer res.Body.Close()

	err = checkResponseStatus(res)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return nil
	}

	var response []models.MonetaryData
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
		return nil
	}

	if mode == "graph" {
		charts.MakeIncomeChart(response)
	}

	tbl, err := tables.MakeTableForMonetaryAggregate(response, accountName)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error building table: %w", err), "Data error")
		return nil
	}
	tbl.Print()
	fmt.Println("")

	return nil
}
