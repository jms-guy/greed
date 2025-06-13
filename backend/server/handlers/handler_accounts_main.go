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

