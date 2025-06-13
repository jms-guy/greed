package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/api/plaidservice"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/encrypt"
	"github.com/jms-guy/greed/models"
)

//Handler populates accounts table in database with account records grabbed from Plaid item ID attached to user
func (app *AppServer) HandlerCreateAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	itemID := chi.URLParam(r, "item-id")
	

	token, err := app.Db.GetAccessToken(ctx, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No item found for user", nil)
			return
		} 
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting access token: %w", err))
		return 
	}

	if token.UserID != id {
		app.respondWithError(w, 403, "UserID does not match item's database record", nil)
		return
	}

	accessTokenbytes, err := encrypt.DecryptAccessToken(token.AccessToken, app.Config.AESKey)
	if err != nil {
		app.respondWithError(w, 500, "Error decrypting access token", err)
		return 
	}

	accessToken := string(accessTokenbytes)

	accounts, reqID, err := plaidservice.GetAccounts(app.PClient, ctx, accessToken)
	if err != nil {
		app.respondWithError(w, 500, "Service Error", fmt.Errorf("error getting accounts from Plaid: %w", err))
		return 
	}

	accRecords := []models.Account{}

	for _, acc := range accounts {

		accSub :=  sql.NullString{}
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
			ID: acc.AccountId,
			Name: acc.Name,
			Type: string(acc.Type),
			Subtype: accSub,
			Mask: accMask,
			OfficialName: accOffName,
			AvailableBalance: accBalAvail,
			CurrentBalance: accBalCur,
			IsoCurrencyCode: curCode,
		}

		dbAcc, err := app.Db.CreateAccount(ctx, params)
		if err != nil {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error creating account record: %w", err))
			return 
		}

		returnAcc := models.Account{
			Id: dbAcc.ID,
			Name: dbAcc.Name,
			Type: dbAcc.Type,
			Subtype: dbAcc.Subtype.String,
			Mask: dbAcc.Mask.String,
			OfficialName: dbAcc.OfficialName.String,
			AvailableBalance: dbAcc.AvailableBalance.String,
			CurrentBalance: dbAcc.CurrentBalance.String,
			IsoCurrencyCode: dbAcc.IsoCurrencyCode.String,
		}

		accRecords = append(accRecords, returnAcc)
	}

	accountsResponse := models.Accounts{
		Accounts: accRecords,
		RequestID: reqID,
	}

	app.respondWithJSON(w, 201, accountsResponse)
}

//Grabs item records for a user from database, returning names and item IDs
func (app *AppServer) HandlerGetItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	items, err := app.Db.GetItemsByUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No item records found", nil)
			return
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting item records: %w", err))
		return
	}
	response := struct {
		Items []models.ItemName `json:"items"`
	}{}
	
	for _, item := range items {
		nickname := ""
		if item.Nickname.Valid {
			nickname = item.Nickname.String
		}
		response.Items = append(response.Items, models.ItemName{
			ItemId: item.ItemID,
			Nickname: nickname,
		})
	}
	app.respondWithJSON(w, 200, response)
}