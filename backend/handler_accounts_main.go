package main

import (
	"encoding/json"
	"net/http"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
)

//Function will get all accounts for user
func (cfg *apiConfig) handlerGetAccountsForUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	//Get accounts for user from database
	accs, err := cfg.db.GetAllAccountsForUser(ctx, id)
	if err != nil {
		respondWithError(w, 500, "Error retrieving accounts for user", err)
		return
	}

	//If no accounts found
	if len(accs) == 0 {
		respondWithError(w, 400, "No accounts found for user", nil)
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

	respondWithJSON(w, 200, accounts)
}
/*
//Function will retrieve single account attached to the given userID and accountID
func (cfg *apiConfig) handlerGetSingleAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Get account data from database based on user ID
	account, err := cfg.db.GetAccount(ctx, database.GetAccountParams{
		ID: acc.ID,
		UserID: id,
	})
	if err != nil {
		respondWithError(w, 500, "Could not retrieve accounts for user", err)
		return
	}

	//If an empty account structure is returned, respond as such
	if account == (database.Account{}) {
		respondWithError(w, 400, "No account found for user", nil)
		return
	}

	//Structure returned database data into return JSON account struct
	response := models.Account{
		ID: account.ID,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
		Name: account.Name,
	}

	respondWithJSON(w, 200, response)
}
*/
//Function will delete an account from the database
func (cfg *apiConfig) handlerDeleteAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Delete account from database based on user ID given
	err := cfg.db.DeleteAccount(ctx, database.DeleteAccountParams{
		ID: acc.ID,
		UserID: id,
	})
	if err != nil {
		respondWithError(w, 500, "Error deleting account from database", err)
		return
	}

	respondWithJSON(w, 200, "Account deleted successfully")
}

//Function will create a new account in the database
func (cfg *apiConfig) handlerCreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := models.AccountDetails{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Couldn't decode parameters", err)
		return
	}

	newAccount, err := cfg.db.CreateAccount(ctx, database.CreateAccountParams{
		ID: uuid.New(),
		Name: params.Name,
		UserID: id,
	})
	if err != nil {
		respondWithError(w, 500, "Error creating account in database", err)
		return
	}

	//Creates the return JSON struct to send back
	account := models.Account{
		ID: newAccount.ID,
		CreatedAt: newAccount.CreatedAt,
		UpdatedAt: newAccount.UpdatedAt,
		Name: newAccount.Name,
		UserID: newAccount.UserID,
	}
	respondWithJSON(w, 201, account)
}
