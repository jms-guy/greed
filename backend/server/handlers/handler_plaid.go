package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/api/plaidservice"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/encrypt"
	"github.com/jms-guy/greed/models"
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

	reqStruct := models.AccessTokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(&reqStruct); err != nil {
		app.respondWithError(w, 400, "Couldn't decode JSON data", err)
		return
	}

	accessToken, err := plaidservice.GetAccessToken(app.PClient, ctx, reqStruct.PublicToken)
	if err != nil {
		app.respondWithError(w, 500, fmt.Sprintf("Error getting access token, Plaid request ID: %s", accessToken.RequestID), fmt.Errorf("reqID: %s, err: %w", accessToken.RequestID, err))
		return 
	}

	encryptedAccessToken, err := encrypt.EncryptAccessToken([]byte(accessToken.AccessToken), app.Config.AESKey)
	if err != nil {
		app.respondWithError(w, 500, "Error encrypting access token", err)
		return
	}

	nickName := sql.NullString{}
	if reqStruct.Nickname != "" {
		nickName.String = reqStruct.Nickname
		nickName.Valid = true
	}

	cursor := sql.NullString{String: "", Valid: true}

	params := database.CreateItemParams{
		ID: accessToken.ItemID,
		UserID: id,
		AccessToken: encryptedAccessToken,
		InstitutionName: accessToken.InstitutionName,
		Nickname: nickName,
		TransactionSyncCursor: cursor,
	}
	_, err = app.Db.CreateItem(ctx, params)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error creating item record: %w", err))
		return
	}

	app.respondWithJSON(w, 201, "Item created")
}