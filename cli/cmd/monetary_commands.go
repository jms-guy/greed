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
func (app *CLIApp) commandGetIncome(cmd *cobra.Command, args []string, mode string) error {
	var account database.Account

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

	incURL := app.Config.Client.BaseURL + "/api/accounts/" + account.ID + "/transactions/monetary"

	res, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", incURL, token, nil)
	})
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer res.Body.Close()

	err = checkResponseStatus(res)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	var response []models.MonetaryData
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
		return err
	}

	if mode == "graph" {
		charts.MakeIncomeChart(response)
	}

	tbl, err := tables.MakeTableForMonetaryAggregate(response, account.Name)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error building table: %w", err), "Data error")
		return err
	}
	tbl.Print()
	fmt.Println("")

	return nil
}
