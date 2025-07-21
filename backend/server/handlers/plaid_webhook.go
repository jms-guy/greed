package handlers

import (
	"encoding/json"
	"net/http"
)

func (app *AppServer) HandlerPlaidWebhook(w http.ResponseWriter, r *http.Request) {
	type Webhook struct{
		WebhookType 			string 	`json:"webhook_type"`
		WebhookCode 			string 	`json:"webhook_code"`
		ItemID					string 	`json:"item_id"`
		InitialUpdateComplete	bool 	`json:"initial_update_complete"`
		HistoricalUpdateComplete bool 	`json:"historical_update_complete"`
	}
	
	request := Webhook{}
	reqErr := json.NewDecoder(r.Body).Decode(&request)	
	if reqErr != nil {
		app.respondWithError(w, 400, "", nil)
		return
	}
}