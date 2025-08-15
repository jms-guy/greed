package cmd

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/cli/internal/utils"
	"github.com/jms-guy/greed/models"
)

// Fetches all transaction records for accounts attached to given item
func (app *CLIApp) commandGetTransactions(args []string) error {
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

	txnsURL := app.Config.Client.BaseURL + "/api/items/" + itemID + "/access/transactions"

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("POST", txnsURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		return err
	}

	var txns []models.Transaction
	if err = json.NewDecoder(resp.Body).Decode(&txns); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	fmt.Println("Creating local records...")

	for _, t := range txns {
		a, err := strconv.ParseFloat(t.Amount, 64)
		if err != nil {
			return fmt.Errorf("error converting string value: %w", err)
		}

		params := database.CreateTransactionParams{
			ID:                      t.Id,
			AccountID:               t.AccountId,
			Amount:                  a,
			IsoCurrencyCode:         sql.NullString{String: t.IsoCurrencyCode, Valid: true},
			Date:                    sql.NullString{String: t.Date.Format("2006-01-02"), Valid: true},
			MerchantName:            sql.NullString{String: t.MerchantName, Valid: true},
			PaymentChannel:          t.PaymentChannel,
			PersonalFinanceCategory: t.PersonalFinanceCategory,
		}

		_, err = app.Config.Db.CreateTransaction(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error creating local record: %w", err)
		}

		fmt.Printf("\r%v", t.Id)
	}

	fmt.Println("\r > Transaction data fetched successfully.")

	return nil
}

// Gets all accounts for item
func (app *CLIApp) commandGetAccounts(args []string) error {
	itemName := args[0]

	itemsURL := app.Config.Client.BaseURL + "/api/items"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

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

	accountsURL := app.Config.Client.BaseURL + "/api/items/" + itemID + "/access/accounts"

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("POST", accountsURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		return err
	}

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

		params := database.UpsertAccountParams{
			ID:               acc.Id,
			CreatedAt:        time.Now().Format("2006-01-02"),
			UpdatedAt:        time.Now().Format("2006-01-02"),
			Name:             acc.Name,
			Type:             acc.Type,
			Subtype:          sql.NullString{String: acc.Subtype, Valid: true},
			Mask:             sql.NullString{String: acc.Mask, Valid: true},
			OfficialName:     sql.NullString{String: acc.OfficialName, Valid: true},
			AvailableBalance: avBalance,
			CurrentBalance:   curBalance,
			IsoCurrencyCode:  sql.NullString{String: acc.IsoCurrencyCode, Valid: true},
			InstitutionName:  sql.NullString{String: itemInst, Valid: true},
			UserID:           creds.User.ID.String(),
		}

		_, err := app.Config.Db.UpsertAccount(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error creating local account record: %w", err)
		}

	}

	fmt.Println(" > Account information fetched successfully.")

	return nil
}

// Rename an item
func (app *CLIApp) commandRenameItem(args []string) error {
	itemCurrent := args[0]
	itemRename := args[1]

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

	renameURL := app.Config.Client.BaseURL + "/api/items/" + itemID + "/name"

	request := models.UpdateItemName{
		Nickname: itemRename,
	}

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("PUT", renameURL, token, request)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		return err
	}

	fmt.Printf("Item name successfully updated from %s to %s\n", itemCurrent, itemRename)
	return nil
}

// Deletes an item record for user
func (app *CLIApp) commandDeleteItem(args []string) error {
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

	pw, err := auth.ReadPassword("Please enter your password > ")
	if err != nil {
		return fmt.Errorf("error getting password: %w", err)
	}

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	user, err := app.Config.Db.GetUser(context.Background(), creds.User.Name)
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

	deleteURL := app.Config.Client.BaseURL + "/api/items/" + itemID

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("DELETE", deleteURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		return err
	}

	params := database.DeleteAccountsParams{
		InstitutionName: sql.NullString{String: itemInst, Valid: true},
		UserID:          user.ID,
	}
	err = app.Config.Db.DeleteAccounts(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error deleting local records: %w", err)
	}

	fmt.Println("Item records deleted successfully")
	return nil
}

// Command resolves Plaid re-authentication issues through Link update flow. Calls sync command immediately afterwards
func (app *CLIApp) commandUpdate(args []string) error {
	itemName := args[0]

	itemsURL := app.Config.Client.BaseURL + "/api/items"
	redirectURL := app.Config.Client.BaseURL + "/link-update-mode"

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

	linkURL := app.Config.Client.BaseURL + "/plaid/get-link-token-update/" + itemID

	linkRes, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("POST", linkURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer linkRes.Body.Close()

	serverErr := parseAndReturnServerError(linkRes)
	if serverErr != nil {
		return serverErr
	}

	var response models.LinkResponse
	if err = json.NewDecoder(linkRes.Body).Decode(&response); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	if response.LinkToken == "" {
		return fmt.Errorf("backend did not return a link token for item update")
	}

	fullBrowserURL := fmt.Sprintf("%s?token=%s", redirectURL, response.LinkToken)
	fmt.Printf("Opening Plaid Link in your browser to update '%s'. Please complete the flow.\n", itemName)
	fmt.Printf("If the browser does not open automatically, please navigate to: %s\n", fullBrowserURL)

	err = utils.OpenLink(app.Config.OperatingSystem, fullBrowserURL)
	if err != nil {
		return err
	}

	_, err = auth.ListenForPlaidCallback()
	if err != nil {
		return err
	}

	err = processWebhookRecords(app, itemID, "ITEM_LOGIN_REQUIRED", "ITEM")
	if err != nil {
		return err
	}
	err = processWebhookRecords(app, itemID, "ITEM_ERROR", "ITEM")
	if err != nil {
		return err
	}
	err = processWebhookRecords(app, itemID, "ITEM_BAD_STATE", "ITEM")
	if err != nil {
		return err
	}
	err = processWebhookRecords(app, itemID, "NEW_ACCOUNTS_AVAILABLE", "ITEM")
	if err != nil {
		return err
	}
	err = processWebhookRecords(app, itemID, "PENDING_DISCONNECT", "ITEM")
	if err != nil {
		return err
	}

	err = app.commandSync([]string{itemName})
	if err != nil {
		return err
	}

	return nil
}

func (app *CLIApp) commandAddItem() error {
	itemsURL := app.Config.Client.BaseURL + "/api/items"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		fmt.Printf("File error: %s\n", err)
		return nil
	}

	linked, err := userFirstTimePlaidLinkHelper(app, creds, itemsURL)
	if err != nil {
		return fmt.Errorf("error linking account : %w", err)
	}

	if !linked {
		fmt.Printf("Linking account to financial institution failed")
		return nil
	}

	return nil
}
