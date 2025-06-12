package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/api/plaidservice"
	"github.com/jms-guy/greed/models"
	"github.com/plaid/plaid-go/v36/plaid"
)

func (app *AppServer) HandlerGetLinkToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	linkToken, err := plaidservice.GetLinkToken(app.PClient, ctx, id.String())
	if err != nil {
		app.respondWithError(w, 500, "Error getting link token from Plaid", err)
		return 
	}

	response := models.LinkResponse{
		LinkToken: linkToken,
	}

	app.respondWithJSON(w, 200, response)
}

func (app *AppServer) HandlerGetAccessToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := plaid.SandboxPublicTokenCreateResponse{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		app.respondWithError(w, 400, "Bad request", err)
		return 
	}

	accessToken, err := plaidservice.GetAccessToken(app.PClient, ctx, request)
	if err != nil {
		app.respondWithError(w, 500, "Error getting access token from Plaid", err)
		return 
	}

		
}