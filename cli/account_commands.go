package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/config"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/models"
)

//Gets all accounts for item
func commandGetAccounts(c *config.Config, args []string) error {
	if len(args) != 1 {
		fmt.Println("Incorrect number of arguments - type --help for more details")
		return nil 
	}

	itemName := args[0]

	itemsURL := c.Client.BaseURL + "/api/items"

	creds, err := auth.GetCreds(c.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	user, err := c.Db.GetUser(context.Background(), creds.User.Name)
	if err != nil {
		return fmt.Errorf("error getting local user record: %w", err)
	}

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

	accountsURL := c.Client.BaseURL + "/api/items/" + itemID + "/access/accounts"

	resp, err := DoWithAutoRefresh(c, func(token string) (*http.Response, error) {
		return c.MakeBasicRequest("POST", accountsURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer resp.Body.Close()

	var response models.Accounts

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	if len(response.Accounts) == 0 {
		fmt.Printf("No accounts found for item: %s\n", itemName)
		return nil
	}
	
	for _, acc := range response.Accounts {
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

		params := database.CreateAccountParams{
			ID: acc.Id,
			CreatedAt: time.Now().Format("2006-01-02"),
			UpdatedAt: time.Now().Format("2006-01-02"),
			Name: acc.Name,
			Type: acc.Type,
			Subtype: sql.NullString{String: acc.Subtype, Valid: true},
			Mask: sql.NullString{String: acc.Mask, Valid: true},
			OfficialName: sql.NullString{String: acc.OfficialName, Valid: true},
			AvailableBalance: avBalance,
			CurrentBalance: curBalance,
			IsoCurrencyCode: sql.NullString{String: acc.IsoCurrencyCode, Valid: true},
			InstitutionName: sql.NullString{String: itemInst, Valid: true},
			UserID: user.ID,
		}

		_, err := c.Db.CreateAccount(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error creating local account record: %w", err)
		}

	}

	tbl := MakeAccountsTable(response.Accounts, itemInst)
	tbl.Print()

	return nil
}

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