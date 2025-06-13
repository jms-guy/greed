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

//Function will get all accounts for user
func (app *AppServer) HandlerGetAccountsForUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	//Get accounts for user from database
	accs, err := app.Db.GetAllAccountsForUser(ctx, id)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error retrieving accounts: %w", err))
		return
	}

	//If no accounts found
	if len(accs) == 0 {
		app.respondWithError(w, 400, "No accounts found for user", nil)
		return
	}

	//Return slice of account structs
	var accounts []models.Account
	for _, account := range accs {
		result := models.Account{
			ID: account.ID,
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
			Name: account.Name,
			UserID: id,
		}
		accounts = append(accounts, result)
	}

	app.respondWithJSON(w, 200, accounts)
}
/*
//Function will retrieve single account attached to the given userID and accountID
func (app *AppServer) handlerGetSingleAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Get account data from database based on user ID
	account, err := app.Db.GetAccount(ctx, database.GetAccountParams{
		ID: acc.ID,
		UserID: id,
	})
	if err != nil {
		app.respondWithError(w, 500, "Could not retrieve accounts for user", err)
		return
	}

	//If an empty account structure is returned, respond as such
	if account == (database.Account{}) {
		app.respondWithError(w, 400, "No account found for user", nil)
		return
	}

	//Structure returned database data into return JSON account struct
	response := models.Account{
		ID: account.ID,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
		Name: account.Name,
	}

	app.respondWithJSON(w, 200, response)
}
*/
//Function will delete an account from the database
func (app *AppServer) HandlerDeleteAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Delete account from database based on user ID given
	err := app.Db.DeleteAccount(ctx, database.DeleteAccountParams{
		ID: acc.ID,
		UserID: id,
	})
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting account: %w", err))
		return
	}

	app.respondWithJSON(w, 200, "Account deleted successfully")
}

//Handler populates accounts table in database with account records grabbed from Plaid item ID attached to user
func (app *AppServer) HandlerCreateAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	token, err := app.Db.GetAccessToken(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No item found for user", nil)
			return
		} 
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting access token: %w", err))
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
