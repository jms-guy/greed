package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/utils"
	"github.com/jms-guy/greed/models"
	"github.com/spf13/cobra"
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

// Goes through Plaid's Link update mode
func linkUpdateModeFlow(app *CLIApp, cmd *cobra.Command, itemID string) error {
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
