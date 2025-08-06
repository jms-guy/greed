package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/cli/internal/tables"
	"github.com/jms-guy/greed/models"
	"net/http"
)

// List accounts for a given item name
func (app *CLIApp) commandListAccounts(args []string) error {

	itemName := args[0]

	itemsURL := app.Config.Client.BaseURL + "/api/items"

	res, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", itemsURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	err = checkResponseStatus(res)
	if err != nil {
		return err
	}

	var itemsResp struct {
		Items []models.ItemName `json:"items"`
	}
	if err = json.NewDecoder(res.Body).Decode(&itemsResp); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	var itemID string
	var itemInst string
	for _, i := range itemsResp.Items {
		if i.Nickname == itemName {
			itemID = i.ItemId
			itemInst = i.InstitutionName
			break
		}
	}

	if itemID == "" {
		fmt.Printf("No item found with name: %s\n", itemName)
		return nil
	}

	accountsURL := app.Config.Client.BaseURL + "/api/items/" + itemID + "/accounts"

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", accountsURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		return err
	}

	var response []models.Account
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	tbl := tables.MakeAccountsTable(response, itemInst)
	tbl.Print()

	return nil
}

// List all accounts for user
func (app *CLIApp) commandListAllAccounts() error {

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	accounts, err := app.Config.Db.GetAllAccounts(context.Background(), creds.User.ID.String())
	if err != nil {
		return fmt.Errorf("error getting local account records: %w", err)
	}

	tbl := tables.MakeAccountsTableAllItems(accounts)
	tbl.Print()

	return nil
}

// Lists account information for a given account name
func (app *CLIApp) commandAccountInfo(args []string) error {

	accountName := args[0]

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	params := database.GetAccountParams{
		Name:   accountName,
		UserID: creds.User.ID.String(),
	}
	account, err := app.Config.Db.GetAccount(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error getting local account record: %w", err)
	}

	tbl := tables.MakeSingleAccountTable(account)
	tbl.Print()

	return nil
}
