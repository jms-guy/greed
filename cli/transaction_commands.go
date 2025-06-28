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

//Get transaction records for given account
//Currently basic, add query arguments, create local txn records
func commandGetTxnsAccount(c *config.Config, args []string) error {
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

	txnsURL := c.Client.BaseURL + "/api/accounts/" + account.ID + "/transactions"

	res, err := DoWithAutoRefresh(c, func(token string) (*http.Response, error) {
		return c.MakeBasicRequest("GET", txnsURL, token, nil)
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
		return nil 
	}

	var txns []models.Transaction
	if err = json.NewDecoder(res.Body).Decode(&txns); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	tbl := MakeTableForTransactions(txns, accountName)
	tbl.Print()

	return nil
}