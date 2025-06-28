package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/config"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/models"
)

//Fetches all transaction records for accounts attached to given item
func commandGetTransactions(c *config.Config, args []string) error {
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
 	for _, i := range itemsResp.Items {
		if i.Nickname == itemName {
			itemID = i.ItemId
			break
		}
	}

	if itemID == "" {
		fmt.Printf("No item found with name: %s\n", itemName)
		return nil 
	}

	txnsURL := c.Client.BaseURL + "/api/items/" + itemID + "/access/transactions"

	resp, err := DoWithAutoRefresh(c, func(token string) (*http.Response, error) {
		return c.MakeBasicRequest("POST", txnsURL, token, nil)
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
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", string(body))
		return nil 
	}

	fmt.Printf("Transaction data successfully generated for %s\n", itemName)
	return nil
}

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

//Rename an item
func commandRenameItem(c *config.Config, args []string) error {
	if len(args) != 2 {
		fmt.Println("Incorrect number of arguments - type --help for more details")
		return nil
	}

	itemCurrent := args[0]
	itemRename := args[1]

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
	for _, i := range itemsResp.Items {
		if i.Nickname == itemCurrent {
			itemID = i.ItemId
			break
		}
	}

	if itemID == "" {
		fmt.Printf("No item found with name: %s\n", itemCurrent)
		return nil 
	}

	renameURL := c.Client.BaseURL + "/api/items/" + itemID + "/name"

	request := models.UpdateItemName{
		Nickname: itemRename,
	}

	resp, err := DoWithAutoRefresh(c, func(token string) (*http.Response, error) {
		return c.MakeBasicRequest("PUT", renameURL, token, request)
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
	}

	fmt.Printf("Item name successfully updated from %s to %s\n", itemCurrent, itemRename)
	return nil
}

//Deletes an item record for user
func commandDeleteItem(c *config.Config, args []string) error {
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

	pw, err := auth.ReadPassword("Please enter your password > ")
	if err != nil {
		return fmt.Errorf("error getting password: %w", err)
	}

	creds, err := auth.GetCreds(c.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	user, err := c.Db.GetUser(context.Background(), creds.User.Name)
	if err != nil {
		return fmt.Errorf("error getting local user record: %w", err)
	}

	err = auth.ValidatePasswordHash(user.HashedPassword, pw)
	if err != nil {
		fmt.Println("Password entered is not correct")
		return nil
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println(" < Upon completion, all records for this item will be permanently deleted. > ")
	fmt.Println(" < If you wish to access this item's data in the future, you will have to link it once again.")
	fmt.Printf(" < Are you sure you want to delete item - %s? (y/n) > \n", itemName)
	for {
		fmt.Print(" > ")
		scanner.Scan()
		if scanner.Text() == "n" {
			fmt.Println("Item deletion aborted.")
			return nil 
		} else if scanner.Text() == "y" {
			break
		} else {
			fmt.Println(" < Please enter either 'y' or 'n' > ")
		}
	}

	deleteURL := c.Client.BaseURL + "/api/items/" + itemID 

	resp, err := DoWithAutoRefresh(c, func(token string) (*http.Response, error) {
		return c.MakeBasicRequest("DELETE", deleteURL, token, nil)
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

	params := database.DeleteAccountsParams{
		InstitutionName: sql.NullString{String: itemInst, Valid: true},
		UserID: user.ID,
	}
	err = c.Db.DeleteAccounts(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error deleting local records: %w", err)
	}

	fmt.Println("Item records deleted successfully")
	return nil
}