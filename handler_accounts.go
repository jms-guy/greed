package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/internal/database"
)

//Structure for account JSON response 
type Account struct {
	ID				uuid.UUID `json:"id"`
	CreatedAt		time.Time `json:"created_at"`
	UpdatedAt		time.Time `json:"updated_at"`
	Balance			string `json:"balance"`
	Goal			string `json:"goal"`
	Currency		string `json:"currency"`
	UserID			uuid.UUID `json:"user_id"`
}

//Function will create a new account in the database
func (cfg *apiConfig) handlerAccountCreate(w http.ResponseWriter, r *http.Request) {
	
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
	var balanceSQL sql.NullString
	if params.Balance == "" {
		balanceSQL.Valid = false
	} else if moneyStringValidation(params.Balance) {
		balanceSQL.Valid = true
		balanceSQL.String = params.Balance
	} else {
		respondWithError(w, 400, "Invalid balance format. Must be in the form of xxx.xx", nil)
		return
	}

	var goalSQL sql.NullString
	if params.Goal == "" {
		goalSQL.Valid = false
	} else if moneyStringValidation(params.Goal) {
		goalSQL.Valid = true
		goalSQL.String = params.Goal
	} else {
		respondWithError(w, 400, "Invalid goal format. Must be in the form of xxx.xx", nil)
		return
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
