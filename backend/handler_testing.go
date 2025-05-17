package main

import (
	"context"
	"net/http"
)

func (cfg *apiConfig) handlerResetTransactions(w http.ResponseWriter, r *http.Request) {
	err := cfg.db.ClearTransactionsTable(context.Background())
	if err != nil {
		respondWithError(w, 500, "Could not clear transactions table", err)
		return
	}
	respondWithJSON(w, 200, "Transactions table cleared successfully")
}

//Function to reset database for dev testing -> users/accounts tables
func (cfg *apiConfig) handlerResetUsersAndAccounts(w http.ResponseWriter, r *http.Request) {
	err := cfg.db.ResetUsers(context.Background())
	if err != nil {
		respondWithError(w, 500, "Could not reset users table", err)
		return
	}

	err = cfg.db.ResetAccounts(context.Background())
	if err != nil {
		respondWithError(w, 500, "Could not reset account_financials table", err)
		return
	}
	respondWithJSON(w, 200, "Users and accounts tables cleared successfully")
}