package main

import (
	"context"
	"encoding/json"
	"net/http"
	"github.com/jms-guy/greed/internal/utils"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/internal/database"
)

//Function will retrieve all accounts attached to the given userID
func (cfg *apiConfig) handlerGetAccount(w http.ResponseWriter, r *http.Request) {
	//Get user ID
	userId := r.PathValue("userid")

	id, err := uuid.Parse(userId)
	if err != nil {
		respondWithError(w, 400, "Error parsing user ID", err)
		return
	}

	//Get account data from database based on user ID
	account, err := cfg.db.GetAccount(context.Background(), id)
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
	response := Account{
		ID: account.ID,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
		Balance: account.Balance.String,
		Goal: account.Goal.String,
		Currency: account.Currency,
	}

	respondWithJSON(w, 200, response)
}

//Function will delete an account from the database
func (cfg *apiConfig) handlerDeleteAccount(w http.ResponseWriter, r *http.Request) {
	//Get user ID
	userId := r.PathValue("userid")

	id, err := uuid.Parse(userId)
	if err != nil {
		respondWithError(w, 400, "Error parsing User ID", err)
		return
	}

	//Delete account from database based on user ID given
	err = cfg.db.DeleteAccount(context.Background(), id)
	if err != nil {
		respondWithError(w, 500, "Error deleting account from database", err)
		return
	}

	respondWithJSON(w, 204, nil)
}

//Function will create a new account in the database
func (cfg *apiConfig) handlerCreateAccount(w http.ResponseWriter, r *http.Request) {
	
	//Parameters that should be present in the received JSON data
	type parameters struct {
		Balance		string `json:"balance,omitempty"`
		Goal		string `json:"goal,omitempty"`
		Currency	string `json:"currency"`
		UserID		uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Couldn't decode parameters", err)
		return
	}

	//Validation on the balance and goal parameters, making sure they are in 
	//proper format, and either null or not null
	balanceSQL, err := utils.CreateNullString(params.Balance)
	if err != nil {
		respondWithError(w, 400, "Invalid string format, expecting (xxx.xx)", nil)
	}

	goalSQL, err := utils.CreateNullString(params.Balance)
	if err != nil {
		respondWithError(w, 400, "Invalid string format, expecting (xxx.xx)", nil)
	}

	//Validation of the currency given, making sure that the given currency
	//is supported in the database ('CAD', 'USD', 'EUR', etc.)
	//If no currency type given for some reason, it's defaulted to CAD
	if params.Currency != "" {
		valid, err := cfg.db.ValidateCurrency(context.Background(), params.Currency)
		if err != nil {
			respondWithError(w, 500, "Error validating currency in database", err)
			return
		}
		if !valid {
			respondWithError(w, 400, "Currency provided is not supported", nil)
			return
		}
	} else {
		params.Currency = "CAD"
	}

	//Creates the account in the database
	newAccount, err := cfg.db.CreateAccount(context.Background(), database.CreateAccountParams{
		ID: uuid.New(),
		Balance: balanceSQL,
		Goal: goalSQL,
		Currency: params.Currency,
		UserID: params.UserID,
	})
	if err != nil {
		respondWithError(w, 500, "Error creating account in database", err)
		return
	}

	//Creates the return JSON struct to send back
	account := Account{
		ID: newAccount.ID,
		CreatedAt: newAccount.CreatedAt,
		UpdatedAt: newAccount.UpdatedAt,
		Balance: newAccount.Balance.String,
		Goal: newAccount.Goal.String,
		Currency: newAccount.Currency,
		UserID: newAccount.UserID,
	}
	respondWithJSON(w, 201, account)
}
