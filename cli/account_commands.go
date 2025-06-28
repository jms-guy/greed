package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/config"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/models"
)

//List accounts for a given item name 
func commandListAccounts(c *config.Config, args []string) error {
	if len(args) != 1 {
		fmt.Println("Incorrect number of arguments - type --help for more details")
		return nil 
	}

	itemName := args[0]

	itemsURL := c.Client.BaseURL + "/api/items"

	res, err := DoWithAutoRefresh(c, func(token string) (*http.Response, error) {
		return c.MakeBasicRequest("GET", itemsURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil 
	}

	if res.StatusCode >= 400 {
		fmt.Printf("Bad request - %s\n", res.Status)
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

	accountsURL := c.Client.BaseURL + "/api/items/" + itemID + "/accounts"

	resp, err := DoWithAutoRefresh(c, func(token string) (*http.Response, error) {
		return c.MakeBasicRequest("GET", accountsURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil 
	}
	if resp.StatusCode >= 400 {
		fmt.Printf("Bad request - %s\n", resp.Status)
		return nil 
	}

	var response []models.Account
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	tbl := MakeAccountsTable(response, itemInst)
	tbl.Print()

	return nil
}

//List all accounts for user
func commandListAllAccounts(c *config.Config, args []string) error {
	if len(args) != 0 {
		fmt.Println("Incorrect number of arguments - type --help for more details")
		return nil 
	}

	creds, err := auth.GetCreds(c.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	user, err := c.Db.GetUser(context.Background(), creds.User.Name)
	if err != nil {
		return fmt.Errorf("error getting local user record: %w", err)
	}

	accounts, err := c.Db.GetAllAccounts(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error getting local account records: %w", err)
	}

	tbl := MakeAccountsTableAllItems(accounts)
	tbl.Print()

	return nil
}

//Lists account information for a given account name
func commandAccountInfo(c *config.Config, args []string) error {
	if len(args) != 1 {
		fmt.Println("Incorrect number of arguments - type --help for more details")
		return nil 
	}

	accountName := args[0]

	creds, err := auth.GetCreds(c.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	user, err := c.Db.GetUser(context.Background(), creds.User.Name)
	if err != nil {
		return fmt.Errorf("error getting local user record: %w", err)
	}

	params := database.GetAccountParams{
		Name: accountName,
		UserID: user.ID,
	}
	account, err := c.Db.GetAccount(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error getting local account record: %w", err)
	}

	tbl := MakeSingleAccountTable(account)
	tbl.Print()

	return nil
}