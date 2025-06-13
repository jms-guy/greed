package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/api/plaidservice"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/encrypt"
	"github.com/jms-guy/greed/models"
	"github.com/plaid/plaid-go/v36/plaid"
)

//Endpoint gets a Link token from Plaid and serves it to the client
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

//Exchanges a received public token with an access token, and stores the Plaid item in database
func (app *AppServer) HandlerGetAccessToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

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

	encryptedAccessToken, err := encrypt.EncryptAccessToken([]byte(accessToken.AccessToken), app.Config.AESKey)
	if err != nil {
		app.respondWithError(w, 500, "Error encrypting access token", err)
		return
	}

	reqID := sql.NullString{}
	if accessToken.RequestID != "" {
		reqID.String = accessToken.RequestID
		reqID.Valid = true
	}

	params := database.CreateItemParams{
		ID: uuid.New(),
		UserID: id,
		ItemID: accessToken.ItemID,
		AccessToken: encryptedAccessToken,
		RequestID: reqID,
	}
	_, err = app.Db.CreateItem(ctx, params)

	app.respondWithJSON(w, 201, "Item created")
}