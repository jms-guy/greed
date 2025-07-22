package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jms-guy/greed/backend/api/plaidservice"
)

//Handler accepts and verifies webhooks from Plaid. Creates database records on what and who the webhook is for.
func (app *AppServer) HandlerPlaidWebhook(w http.ResponseWriter, r *http.Request) {
	type Webhook struct{
		WebhookType 			string 	`json:"webhook_type"`
		WebhookCode 			string 	`json:"webhook_code"`
		ItemID					string 	`json:"item_id"`
		InitialUpdateComplete	bool 	`json:"initial_update_complete"`
		HistoricalUpdateComplete bool 	`json:"historical_update_complete"`
	}

	ctx := r.Context()
	
	request := Webhook{}
	reqErr := json.NewDecoder(r.Body).Decode(&request)	
	if reqErr != nil {
		app.respondWithError(w, 400, "", nil)
		return
	}

	tokenString := r.Header.Get("plaid-verification")


	plaidService := app.PService.(*plaidservice.PlaidService)	//Type assertion for use in auth method
	err := app.Auth.VerifyPlaidJWT(plaidService, ctx, tokenString)
	if err != nil {
		app.respondWithError(w, 400, "", fmt.Errorf("error verifying plaid JWT: %w", err))
		return
	}

	
}