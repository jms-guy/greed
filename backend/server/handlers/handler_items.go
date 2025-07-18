package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/api/plaidservice"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
)

//Handler populates accounts table in database with account records grabbed from Plaid item ID attached to user
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

	accounts, reqID, err := plaidservice.GetAccounts(app.PClient, ctx, accessToken)
	if err != nil {
		app.respondWithError(w, 500, "Service Error", fmt.Errorf("plaid request id: %s, error getting accounts from Plaid: %w", reqID, err))
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
			ItemID: itemID,
			UserID: id,
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
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	items, err := app.Db.GetItemsByUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No items found for user", nil)
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
			ItemId: item.ID,
			Nickname: nickname,
			InstitutionName: item.InstitutionName,
		})
	}
	app.respondWithJSON(w, 200, response)
}

//Updates item's nickname field in database
func (app *AppServer) HandlerUpdateItemName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	request := models.UpdateItemName{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		app.respondWithError(w, 400, "Bad request data", err)
		return 
	}

	itemID := chi.URLParam(r, "item-id")

	params := database.UpdateNicknameParams{
		Nickname: sql.NullString{String: request.Nickname, Valid: true},
		ID: itemID,
		UserID: id,
	}

	err := app.Db.UpdateNickname(ctx, params)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error updating item name: %w", err))
		return 
	}

	app.respondWithJSON(w, 200, "Item name updated successfully")
}


//Endpoint deletes a user's item from database
func (app *AppServer) HandlerDeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	itemID := chi.URLParam(r, "item-id")

	params := database.DeleteItemParams{
		ID: itemID,
		UserID: id,
	}

	err := app.Db.DeleteItem(ctx, params)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting item record: %w", err))
		return 
	}

	app.respondWithJSON(w, 200, "Item deleted successfully")
}

//Gets accounts only for a user's specific item
func (app *AppServer) HandlerGetAccountsForItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	itemID := chi.URLParam(r, "item-id")

	accs, err := app.Db.GetAccountsForItem(ctx, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No accounts found for item", nil)
			return 
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting accounts for item: %w", err))
		return 
	}

	//Return slice of account structs
	var accounts []models.Account
	for _, account := range accs {
		result := models.Account{
			Id: account.ID,
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
			Name: account.Name,
			Type: account.Type,
			Subtype: account.Subtype.String,
			Mask: account.Mask.String,
			OfficialName: account.OfficialName.String,
			AvailableBalance: account.AvailableBalance.String,
			CurrentBalance: account.CurrentBalance.String,
			IsoCurrencyCode: account.IsoCurrencyCode.String,
		}
		accounts = append(accounts, result)
	}
	
	app.respondWithJSON(w, 200, accounts)
}

//Function to populate database with transaction data for a given item
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
		ID: itemID,
		AccessToken: accessToken,
	}
	cursor, err := app.Db.GetCursor(ctx, params)
	if err != nil && err != sql.ErrNoRows{
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting cursor: %w", err))
		return 
	}

	added, modified, removed, nextCursor, reqID, err := plaidservice.GetTransactions(app.PClient, ctx, accessToken, cursor.String)
	if err != nil {
		app.respondWithError(w, 500, "Service error", fmt.Errorf("plaid request id: %s, error getting transaction data: %w", reqID, err))
		return 
	}

	err = ApplyTransactionUpdates(app, ctx, added, modified, removed, nextCursor, itemID)
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
			Id: t.ID,
			AccountId: t.AccountID,
			Amount: t.Amount,
			IsoCurrencyCode: t.IsoCurrencyCode.String,
			Date: t.Date.Time,
			MerchantName: t.MerchantName.String,
			PaymentChannel: t.PaymentChannel,
			PersonalFinanceCategory: t.PersonalFinanceCategory,
		}
		response = append(response, newT)
	}

	app.respondWithJSON(w, 200, response)
}