package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
	"github.com/plaid/plaid-go/v36/plaid"
)

func (app *AppServer) HandlerGetSandboxToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	accessToken, err := app.PSandbox.GetSandboxToken(ctx)
	if err != nil {
		app.respondWithError(w, 500, "Service error", fmt.Errorf("error getting sandbox token: %w", err))
		return
	}

	encryptedAccessToken, err := app.Encryptor.EncryptAccessToken([]byte(accessToken.AccessToken), app.Config.AESKey)
	if err != nil {
		app.respondWithError(w, 500, "Error encrypting access token", err)
		return
	}

	nickName := sql.NullString{String: "Test", Valid: true}

	cursor := sql.NullString{String: "", Valid: true}

	params := database.CreateItemParams{
		ID:                    accessToken.ItemId,
		UserID:                id,
		AccessToken:           encryptedAccessToken,
		InstitutionName:       "Plaid Banking",
		Nickname:              nickName,
		TransactionSyncCursor: cursor,
	}

	_, err = app.Db.CreateItem(ctx, params)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error creating item record: %w", err))
		return
	}

	app.respondWithJSON(w, 201, "Item created")
}

// Endpoint gets a Link token from Plaid and serves it to the client
func (app *AppServer) HandlerGetLinkToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	linkToken, err := app.PService.GetLinkToken(ctx, id.String(), app.Config.PlaidWebhookURL)
	if err != nil {
		if plaidErr, ok := err.(plaid.GenericOpenAPIError); ok {
			fmt.Printf("Plaid error: %s\n", string(plaidErr.Body()))
		} else {
			fmt.Println("Error:", err.Error())
		}
		app.respondWithError(w, 500, "Error getting link token from Plaid", err)
		return
	}

	response := models.LinkResponse{
		LinkToken: linkToken,
	}

	app.respondWithJSON(w, 200, response)
}

// Gets a Link token from Plaid with user's Access token, providing Update mode re-authentication
func (app *AppServer) HandlerGetLinkTokenForUpdateMode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	tokenValue := ctx.Value(accessTokenKey)
	accessToken, ok := tokenValue.(string)
	if !ok {
		app.respondWithError(w, 400, "Bad access token in context", nil)
		return
	}

	linkToken, err := app.PService.GetLinkTokenForUpdateMode(ctx, id.String(), accessToken, app.Config.PlaidWebhookURL)
	if err != nil {
		if plaidErr, ok := err.(plaid.GenericOpenAPIError); ok {
			fmt.Printf("Plaid error: %s\n", string(plaidErr.Body()))
		} else {
			fmt.Println("Error:", err.Error())
		}
		app.respondWithError(w, 500, "Service error", fmt.Errorf("error getting link token: %w", err))
		return
	}

	response := models.LinkResponse{
		LinkToken: linkToken,
	}

	app.respondWithJSON(w, 200, response)
}

// Exchanges a received public token with an access token, and stores the Plaid item in database
func (app *AppServer) HandlerGetAccessToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	reqStruct := models.AccessTokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(&reqStruct); err != nil {
		app.respondWithError(w, 400, "Couldn't decode JSON data", err)
		return
	}

	accessToken, err := app.PService.GetAccessToken(ctx, reqStruct.PublicToken)
	if err != nil {
		app.respondWithError(w, 500, fmt.Sprintf("Error getting access token, Plaid request ID: %s", accessToken.RequestID), fmt.Errorf("reqID: %s, err: %w", accessToken.RequestID, err))
		return
	}

	encryptedAccessToken, err := app.Encryptor.EncryptAccessToken([]byte(accessToken.AccessToken), app.Config.AESKey)
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
		ID:                    accessToken.ItemID,
		UserID:                id,
		AccessToken:           encryptedAccessToken,
		InstitutionName:       accessToken.InstitutionName,
		Nickname:              nickName,
		TransactionSyncCursor: cursor,
	}
	_, err = app.Db.CreateItem(ctx, params)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error creating item record: %w", err))
		return
	}

	app.respondWithJSON(w, 201, "Item created")
}

// Handler populates accounts table in database with account records grabbed from Plaid item ID attached to user
func (app *AppServer) HandlerCreateAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	itemID := chi.URLParam(r, "item-id")

	tokenValue := ctx.Value(accessTokenKey)
	accessToken, ok := tokenValue.(string)
	if !ok {
		app.respondWithError(w, 400, "Bad access token in context", nil)
		return
	}

	accounts, reqID, err := app.PService.GetAccounts(ctx, accessToken)
	if err != nil {
		app.respondWithError(w, 500, "Service Error", fmt.Errorf("plaid request id: %s, error getting accounts from Plaid: %w", reqID, err))
		return
	}

	accRecords := []models.Account{}

	for _, acc := range accounts {

		accSub := sql.NullString{}
		if acc.Subtype.IsSet() {
			accSub.String = string(*acc.Subtype.Get().Ptr())
			accSub.Valid = true
		}

		accMask := sql.NullString{}
		if acc.Mask.IsSet() {
			accMask.String = *acc.Mask.Get()
			accMask.Valid = true
		}

		accOffName := sql.NullString{}
		if acc.OfficialName.IsSet() {
			accOffName.String = *acc.OfficialName.Get()
			accOffName.Valid = true
		}

		accBalAvail := sql.NullString{}
		if acc.Balances.Available.IsSet() {
			accBalAvail.String = fmt.Sprintf("%.2f", *acc.Balances.Available.Get())
			accBalAvail.Valid = true
		}

		accBalCur := sql.NullString{}
		if acc.Balances.Current.IsSet() {
			accBalCur.String = fmt.Sprintf("%.2f", *acc.Balances.Current.Get())
			accBalCur.Valid = true
		}

		curCode := sql.NullString{}
		if acc.Balances.IsoCurrencyCode.IsSet() {
			curCode.String = acc.Balances.GetIsoCurrencyCode()
			curCode.Valid = true
		}

		params := database.CreateAccountParams{
			ID:               acc.AccountId,
			Name:             acc.Name,
			Type:             string(acc.Type),
			Subtype:          accSub,
			Mask:             accMask,
			OfficialName:     accOffName,
			AvailableBalance: accBalAvail,
			CurrentBalance:   accBalCur,
			IsoCurrencyCode:  curCode,
			ItemID:           itemID,
			UserID:           id,
		}

		dbAcc, err := app.Db.CreateAccount(ctx, params)
		if err != nil {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error creating account record: %w", err))
			return
		}

		returnAcc := models.Account{
			Id:               dbAcc.ID,
			Name:             dbAcc.Name,
			Type:             dbAcc.Type,
			Subtype:          dbAcc.Subtype.String,
			Mask:             dbAcc.Mask.String,
			OfficialName:     dbAcc.OfficialName.String,
			AvailableBalance: dbAcc.AvailableBalance.String,
			CurrentBalance:   dbAcc.CurrentBalance.String,
			IsoCurrencyCode:  dbAcc.IsoCurrencyCode.String,
		}

		accRecords = append(accRecords, returnAcc)
	}

	accountsResponse := models.Accounts{
		Accounts:  accRecords,
		RequestID: reqID,
	}

	app.respondWithJSON(w, 201, accountsResponse)
}

// Function to populate database with transaction data for a given item
func (app *AppServer) HandlerSyncTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tokenValue := ctx.Value(accessTokenKey)
	accessToken, ok := tokenValue.(string)
	if !ok {
		app.respondWithError(w, 400, "Bad access token in context", nil)
		return
	}

	itemID := chi.URLParam(r, "item-id")

	params := database.GetCursorParams{
		ID:          itemID,
		AccessToken: accessToken,
	}
	cursor, err := app.Db.GetCursor(ctx, params)
	if err != nil && err != sql.ErrNoRows {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting cursor: %w", err))
		return
	}

	added, modified, removed, nextCursor, reqID, err := app.PService.GetTransactions(ctx, accessToken, cursor.String)
	if err != nil {
		app.respondWithError(w, 500, "Service error", fmt.Errorf("plaid request id: %s, error getting transaction data: %w", reqID, err))
		return
	}

	err = app.TxnUpdater.ApplyTransactionUpdates(ctx, added, modified, removed, nextCursor, itemID)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error completing database txn on transactional data: %w", err))
		return
	}

	item, err := app.Db.GetItemByID(ctx, itemID)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting item item record: %w", err))
		return
	}

	txns, err := app.Db.GetTransactionsForUser(ctx, item.UserID)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting transaction records: %w", err))
		return
	}

	var response []models.Transaction
	for _, t := range txns {
		newT := models.Transaction{
			Id:                      t.ID,
			AccountId:               t.AccountID,
			Amount:                  t.Amount,
			IsoCurrencyCode:         t.IsoCurrencyCode.String,
			Date:                    t.Date.Time,
			MerchantName:            t.MerchantName.String,
			PaymentChannel:          t.PaymentChannel,
			PersonalFinanceCategory: t.PersonalFinanceCategory,
		}
		response = append(response, newT)
	}

	app.respondWithJSON(w, 200, response)
}

// Updates the balances of all accounts associated with an item
func (app *AppServer) HandlerUpdateBalances(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tokenValue := ctx.Value(accessTokenKey)
	accessToken, ok := tokenValue.(string)
	if !ok {
		app.respondWithError(w, 400, "Bad access token in context", nil)
		return
	}

	accs, reqID, err := app.PService.GetBalances(ctx, accessToken)
	if err != nil {
		if plaidErr, ok := err.(plaid.GenericOpenAPIError); ok {
			fmt.Printf("Plaid error: %s\n", string(plaidErr.Body()))
		} else {
			fmt.Println("Error:", err.Error())
		}
		app.respondWithError(w, 500, "Service error", fmt.Errorf("plaid request id: %s, error getting updated account balances: %w", reqID, err))
		return
	}

	responseAccounts := models.Accounts{}
	for _, acc := range accs.Accounts {

		accBalAvail := sql.NullString{}
		if acc.Balances.Available.IsSet() {
			accBalAvail.String = fmt.Sprintf("%.2f", *acc.Balances.Available.Get())
			accBalAvail.Valid = true
		}

		accBalCur := sql.NullString{}
		if acc.Balances.Current.IsSet() {
			accBalCur.String = fmt.Sprintf("%.2f", *acc.Balances.Current.Get())
			accBalCur.Valid = true
		}

		params := database.UpdateBalancesParams{
			AvailableBalance: accBalAvail,
			CurrentBalance:   accBalCur,
			ID:               acc.AccountId,
			ItemID:           accs.Item.ItemId,
		}

		_, err := app.Db.UpdateBalances(ctx, params)
		if err != nil {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error updating account record: %w", err))
			return
		}

		updatedRecord := models.Account{
			Id:               acc.AccountId,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			Name:             acc.Name,
			Type:             string(acc.Type),
			Subtype:          string(acc.GetSubtype()),
			Mask:             acc.GetMask(),
			OfficialName:     acc.GetOfficialName(),
			AvailableBalance: accBalAvail.String,
			CurrentBalance:   accBalCur.String,
			IsoCurrencyCode:  acc.Balances.GetIsoCurrencyCode(),
		}

		responseAccounts.Accounts = append(responseAccounts.Accounts, updatedRecord)
	}

	app.respondWithJSON(w, 200, responseAccounts)
}
