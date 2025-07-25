package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jms-guy/greed/models"
)

//Gets list of items from server, finds a specific item based on name and returns the ID
func findItemHelper(app *CLIApp, itemName, itemsURL string) (string, error) {
	res, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", itemsURL, token, nil)
	})
	if err != nil {
		return "", fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	err = checkResponseStatus(res)
	if err != nil {
		return "", err
	}

	var itemsResp struct {
		Items []models.ItemName `json:"items"`
	}
	if err = json.NewDecoder(res.Body).Decode(&itemsResp); err != nil {
		return "", fmt.Errorf("error decoding response data: %w", err)
	}

	var itemID string
 	for _, i := range itemsResp.Items {
		if i.Nickname == itemName {
			itemID = i.ItemId
			break
		}
	}

	if itemID == "" {
		return "", fmt.Errorf("no item found with name: %s", itemName)
	}

	return itemID, nil
}

//Slightly more in depth status code handling for sync command http responses
func parseAndReturnServerError(resp *http.Response, ) error {
	if resp.StatusCode != http.StatusOK {
		var serverErr map[string]string
		if decodeErr := json.NewDecoder(resp.Body).Decode(&serverErr); decodeErr == nil && serverErr["error"] != "" {
			return fmt.Errorf("server error syncing transactions (status %d): %s", resp.StatusCode, serverErr["error"])
		}
		return fmt.Errorf("server returned non-OK status for transactions: %s", resp.Status)
	}

	return nil
}

//Makes server request to process webhooks of a certain type
func processWebhookRecords(app *CLIApp, itemID, webhookCode, webhookType string) error {
	webhookURL := app.Config.Client.BaseURL + "/api/items/webhook-records"

	request := models.ProcessWebhook{
		ItemID: itemID,
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