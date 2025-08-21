package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/cli/internal/tables"
	"github.com/jms-guy/greed/models"
	"github.com/spf13/cobra"
)

// List accounts for a given item name
func (app *CLIApp) commandListAccounts(cmd *cobra.Command, args []string) error {
	itemName := args[0]

	item, err := getItemFromServer(app, itemName)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting item")
		return err
	}

	accountsURL := app.Config.Client.BaseURL + "/api/items/" + item.ItemId + "/accounts"

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", accountsURL, token, nil)
	})
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	var response []models.Account
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
		return err
	}

	tbl := tables.MakeAccountsTable(response, item.InstitutionName)
	tbl.Print()

	return nil
}

// List all accounts for user
func (app *CLIApp) commandListAllAccounts(cmd *cobra.Command) error {
	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting credentials")
		return err
	}

	accounts, err := app.Config.Db.GetAllAccounts(context.Background(), creds.User.ID.String())
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error getitng local account records: %w", err), "Local database error")
		return err
	}

	tbl := tables.MakeAccountsTableAllItems(accounts)
	tbl.Print()

	return nil
}

// Lists account information for a given account name
func (app *CLIApp) commandAccountInfo(cmd *cobra.Command, args []string) error {
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

	tbl := tables.MakeSingleAccountTable(account)
	tbl.Print()

	return nil
}
