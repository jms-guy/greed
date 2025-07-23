package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
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
		app.respondWithError(w, 400, "Bad request", nil)
		return
	}

	tokenString := r.Header.Get("plaid-verification")


	
	err := app.Auth.VerifyPlaidJWT(app.PService, ctx, tokenString)
	if err != nil {
		app.respondWithError(w, 401, "Unauthorized", fmt.Errorf("error verifying plaid JWT: %w", err))
		return
	}

	item, err := app.Db.GetItemByID(ctx, request.ItemID)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 404, "Not Found", fmt.Errorf("itemID in webhook request does not exist"))
			return
		}
		app.respondWithError(w, 500, "Internal Server Error", fmt.Errorf("error getting item based on webhook itemID: %w", err))
		return
	}

	params := database.CreatePlaidWebhookRecordParams{
		ID: uuid.New(),
		WebhookType: request.WebhookType,
		WebhookCode: request.WebhookCode,
		UserID: item.UserID,
		ItemID: request.ItemID,
	}

	_, err = app.Db.CreatePlaidWebhookRecord(ctx, params)
	if err != nil {
		app.respondWithError(w, 500, "Internal Server Error", fmt.Errorf("error creating webhook record: %w", err))
		return
	}

	app.respondWithJSON(w, 200, "")
}