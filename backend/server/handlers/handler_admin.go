package handlers

import (
	"context"
	"net/http"
	"fmt"
)

//Function returns list of all user names in database
func (app *AppServer) HandlerGetListOfUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := app.Db.GetAllUsers(ctx)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting users: %w", err))
		return
	}

	app.respondWithJSON(w, 200, users)
}

//Function to reset plaid_items database table
func (app *AppServer) HandlerResetItems(w http.ResponseWriter, r *http.Request) {
	err := app.Db.ResetItems(context.Background())
	if err != nil {
		app.respondWithError(w, 500, "Could not reset plaid_items table", err)
		return 
	}
	app.respondWithJSON(w, 200, "Items table cleared successfully")
}

//Function to reset database accounts table
func (app *AppServer) HandlerResetAccounts(w http.ResponseWriter, r *http.Request) {
	err := app.Db.ResetAccounts(context.Background())
	if err != nil {
		app.respondWithError(w, 500, "Could not reset account_financials table", err)
		return
	}
	app.respondWithJSON(w, 200, "Accounts table cleared successfully")
}

//Function will reset database's transactions table
func (app *AppServer) HandlerResetTransactions(w http.ResponseWriter, r *http.Request) {
	err := app.Db.ClearTransactionsTable(context.Background())
	if err != nil {
		app.respondWithError(w, 500, "Could not clear transactions table", err)
		return
	}
	app.respondWithJSON(w, 200, "Transactions table cleared successfully")
}

//Function to reset database for dev testing -> users/accounts tables
func (app *AppServer) HandlerResetUsers(w http.ResponseWriter, r *http.Request) {
	err := app.Db.ResetUsers(context.Background())
	if err != nil {
		app.respondWithError(w, 500, "Could not reset users table", err)
		return
	}
	app.respondWithJSON(w, 200, "Users table cleared successfully")
}