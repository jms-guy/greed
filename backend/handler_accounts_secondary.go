package main

import (
	"context"
	"encoding/json"
	"net/http"
	"github.com/jms-guy/greed/internal/utils"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/internal/database"
)

//Function will update an account's currency field in the database, based on given account ID
func (cfg *apiConfig) handlerUpdateCurrency(w http.ResponseWriter, r *http.Request) {
	//Request parameters needed
	type parameters struct {
		Currency		string `json:"currency"`
	}

	//Get account ID
	accountId := r.PathValue("accountid")

	id, err := uuid.Parse(accountId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Decode request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Couldn't decode JSON parameters", err)
		return
	}

	//Validate that the currency provided is supported in the database
	valid, err := cfg.db.ValidateCurrency(context.Background(), params.Currency)
	if err != nil {
		respondWithError(w, 500, "Error validating currency in database", err)
		return
	}
	if !valid {
		respondWithError(w, 400, "Currency provided is not supported", nil)
		return
	}

	//Update the database
	err = cfg.db.UpdateCurrency(context.Background(), database.UpdateCurrencyParams{
		Currency: params.Currency,
		ID: id,
	})
	if err != nil {
		respondWithError(w, 500, "Error updating account currency setting", err)
		return
	}

	respondWithJSON(w, 200, Account{Currency: params.Currency})
}

//Function will update an account's goal field in the database based on a given account ID
func (cfg *apiConfig) handlerUpdateGoal(w http.ResponseWriter, r *http.Request) {
	//Request parameters needed
	type parameters struct {
		Goal		string `json:"goal"`
	}

	//Get account ID
	accountId := r.PathValue("accountid")

	id, err := uuid.Parse(accountId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Decode request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Coudln't decode JSON parameters", err)
		return
	}

	//Turn goal string into type NullString for database
	goalSQL, err := utils.CreateMoneyNullString(params.Goal)
	if err != nil {
		respondWithError(w, 400, "Invalid string format, expecting (xxx.xx)", err)
		return
	}

	//Update database
	err = cfg.db.UpdateGoal(context.Background(), database.UpdateGoalParams{
		Goal: goalSQL,
		ID: id,
	})
	if err != nil {
		respondWithError(w, 500, "Error updating account goal", err)
		return
	}

	//Respond with Goal data
	respondWithJSON(w, 200, Account{Goal: params.Goal})
}

func (cfg *apiConfig) handlerUpdateBalance(w http.ResponseWriter, r *http.Request) {
	//Request parameters needed
	type parameters struct {
		Balance		string`json:"balance"`
	}

	//Get account ID 
	accountId := r.PathValue("accountid")
	
	id, err := uuid.Parse(accountId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Decode request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Coudln't decode JSON parameters", err)
		return
	}

	//Turn request balance string into type NullString
	balanceSQL, err := utils.CreateMoneyNullString(params.Balance)
	if err != nil {
		respondWithError(w, 400, "Invalid string format, expecting (xxx.xx)", nil)
		return
	}

	//Update database
	err = cfg.db.UpdateBalance(context.Background(), database.UpdateBalanceParams{
		Balance: balanceSQL,
		ID: id,
	})
	if err != nil {
		respondWithError(w, 500, "Error updating account balance", err)
		return
	}

	//Respond with Balance data
	respondWithJSON(w, 200, Account{Balance: params.Balance})
}