package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/cli/internal/utils"
	"github.com/jms-guy/greed/models"
)

// Gets list of items from server, finds a specific item based on name and returns the ID and institution
func findItemHelper(app *CLIApp, itemName, itemsURL string) (string, string, error) {
	res, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", itemsURL, token, nil)
	})
	if err != nil {
		return "", "", fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	err = checkResponseStatus(res)
	if err != nil {
		return "", "", err
	}

	var itemsResp struct {
		Items []models.ItemName `json:"items"`
	}
	if err = json.NewDecoder(res.Body).Decode(&itemsResp); err != nil {
		return "", "", fmt.Errorf("error decoding response data: %w", err)
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
		return "", "", fmt.Errorf("no item found with name: %s", itemName)
	}

	return itemID, itemInst, nil
}

// Slightly more in depth status code handling for sync command http responses
func parseAndReturnServerError(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		var serverErr map[string]string
		if decodeErr := json.NewDecoder(resp.Body).Decode(&serverErr); decodeErr == nil && serverErr["error"] != "" {
			return fmt.Errorf("server error syncing transactions (status %d): %s", resp.StatusCode, serverErr["error"])
		}
		return fmt.Errorf("server returned non-OK status for transactions: %s", resp.Status)
	}

	return nil
}

// Makes server request to process webhooks of a certain type
func processWebhookRecords(app *CLIApp, itemID, webhookCode, webhookType string) error {
	webhookURL := app.Config.Client.BaseURL + "/api/items/webhook-records"

	request := models.ProcessWebhook{
		ItemID:      itemID,
		WebhookCode: webhookCode,
		WebhookType: webhookType,
	}

	res, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("PUT", webhookURL, token, request)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	err = checkResponseStatus(res)
	if err != nil {
		return err
	}

	return nil
}

// Function updates user's account records with fresh balance data obtained from Plaid
func syncAccountBalances(app *CLIApp, creds models.Credentials, itemID, itemInst string) error {
	updateBalanceURL := app.Config.Client.BaseURL + "/api/items/" + itemID + "/access/balances"

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("PUT", updateBalanceURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer resp.Body.Close()

	var accUpdates models.Accounts

	serverErr := parseAndReturnServerError(resp)
	if serverErr != nil {
		return serverErr
	}
	if err = json.NewDecoder(resp.Body).Decode(&accUpdates); err != nil {
		if err == io.EOF {
			return fmt.Errorf("decoding err: %w", err)
		}
		return fmt.Errorf("decoding err: %w", err)
	}

	fmt.Println(" > Syncing account balances...")

	for _, acc := range accUpdates.Accounts {
		avBalance := sql.NullFloat64{}
		if acc.AvailableBalance != "" {
			avBal, err := strconv.ParseFloat(acc.AvailableBalance, 64)
			if err != nil {
				return fmt.Errorf("error converting string value: %w", err)
			}
			avBalance.Float64 = avBal
			avBalance.Valid = true
		}

		curBalance := sql.NullFloat64{}
		if acc.CurrentBalance != "" {
			curBal, err := strconv.ParseFloat(acc.CurrentBalance, 64)
			if err != nil {
				return fmt.Errorf("error converting string value: %w", err)
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
		_, err = app.Config.Db.UpsertAccount(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error upserting account record: %w", err)
		}
	}

	fmt.Println(" > Account balances synced successfully.")
	return nil
}

// Function deletes local transaction records for an item, replacing them with updated data fetched
// from Plaid and updated server-side
func syncTransactionData(app *CLIApp, creds models.Credentials, itemID string) error {
	syncTxnsURL := app.Config.Client.BaseURL + "/api/items/" + itemID + "/access/transactions"

	response, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("POST", syncTxnsURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer response.Body.Close()

	serverErr := parseAndReturnServerError(response)
	if serverErr != nil {
		return serverErr
	}

	var txns []models.Transaction
	if err = json.NewDecoder(response.Body).Decode(&txns); err != nil {
		if err == io.EOF {
			return fmt.Errorf("decoding err: %w", err)
		}
		return fmt.Errorf("decoding err: %w", err)
	}

	fmt.Println(" > Syncing transaction records...")

	err = app.Config.Db.DeleteTransactions(context.Background(), creds.User.ID.String())
	if err != nil {
		return fmt.Errorf("error clearing local records: %w", err)
	}

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
			return fmt.Errorf("error creating local records: %w", err)
		}

		fmt.Printf("\r > %v", t.Id)
	}

	fmt.Println("\r > Transaction records synced successfully.")
	return nil
}

// Goes through Plaid's Link update mode
func linkUpdateModeFlow(app *CLIApp, itemID string) error {
	linkURL := app.Config.Client.BaseURL + "/plaid/get-link-token-update/" + itemID
	redirectURL := app.Config.Client.BaseURL + "/link-update-mode"

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
		return fmt.Errorf("decoding err: %w", err)
	}

	if response.LinkToken == "" {
		return fmt.Errorf("backend did not return a link token for item update")
	}

	fullBrowserURL := fmt.Sprintf("%s?token=%s", redirectURL, response.LinkToken)
	fmt.Printf("Opening Plaid Link in your browser to update item with ID '%s'. Please complete the flow.\n", itemID)
	fmt.Printf("If the browser does not open automatically, please navigate to: %s\n", fullBrowserURL)

	err = utils.OpenLink(app.Config.OperatingSystem, fullBrowserURL)
	if err != nil {
		return err
	}

	_, err = auth.ListenForPlaidCallback()
	if err != nil {
		return err
	}

	webhookCodes := []string{"ITEM_LOGIN_REQUIRED", "ITEM_ERROR", "ITEM_BAD_STATE", "NEW_ACCOUNTS_AVAILABLE", "PENDING_DISCONNECT", "ERROR"}

	for _, code := range webhookCodes {
		err = processWebhookRecords(app, itemID, code, "ITEM")
		if err != nil {
			return err
		}
	}

	return nil
}
