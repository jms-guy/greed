package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jms-guy/greed/models"
)

// Retrieves an item record from server database based on given item name
func getItemFromServer(app *CLIApp, itemName string) (models.ItemName, error) {
	var itemFound models.ItemName
	itemsURL := app.Config.Client.BaseURL + "/api/items"

	res, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", itemsURL, token, nil)
	})
	if err != nil {
		return itemFound, fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	err = checkResponseStatus(res)
	if err != nil {
		return itemFound, err
	}

	var itemsResp struct {
		Items []models.ItemName `json:"items"`
	}
	if err = json.NewDecoder(res.Body).Decode(&itemsResp); err != nil {
		return itemFound, fmt.Errorf("decoding err: %w", err)
	}

	for _, i := range itemsResp.Items {
		if i.Nickname == itemName {
			itemFound = i
			break
		}
	}

	if itemFound.ItemId == "" {
		return itemFound, fmt.Errorf("no item found")
	}

	return itemFound, nil
}
