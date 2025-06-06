package main

import (
	"context"
	"net/http"
)

//Function to reset database accounts table
func (app *AppServer) handlerResetAccounts(w http.ResponseWriter, r *http.Request) {
	err := app.db.ResetAccounts(context.Background())
	if err != nil {
		respondWithError(w, 500, "Could not reset account_financials table", err)
		return
	}
	respondWithJSON(w, 200, "Accounts table cleared successfully")
}

//Function will reset database's transactions table
func (app *AppServer) handlerResetTransactions(w http.ResponseWriter, r *http.Request) {
	err := app.db.ClearTransactionsTable(context.Background())
	if err != nil {
		respondWithError(w, 500, "Could not clear transactions table", err)
		return
	}
	respondWithJSON(w, 200, "Transactions table cleared successfully")
}

//Function to reset database for dev testing -> users/accounts tables
func (app *AppServer) handlerResetUsers(w http.ResponseWriter, r *http.Request) {
	err := app.db.ResetUsers(context.Background())
	if err != nil {
		respondWithError(w, 500, "Could not reset users table", err)
		return
	}
	respondWithJSON(w, 200, "Users table cleared successfully")
}