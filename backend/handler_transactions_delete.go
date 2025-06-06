package main

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/jms-guy/greed/backend/internal/database"
)

//Function deletes all transaction records for an account of a given category
func (app *AppServer) handlerDeleteTransactionsOfCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	cat := chi.URLParam(r, "category")
	
	if cat == "" {
		respondWithError(w, 400, "Bad category given", nil)
		return
	}

	//Delete transactions from database
	err := app.db.DeleteTransactionsOfCategory(ctx, database.DeleteTransactionsOfCategoryParams{
		AccountID: acc.ID,
		Category: cat,
	})
	if err != nil {
		respondWithError(w, 500, "Error deleting transactions", err)
		return
	}

	respondWithJSON(w, 200, "Transactions deleted successfully")
}

//Function deletes all transaction records for an account based on a given transaction_type
//(credit, debit, transfer)
func (app *AppServer) handlerDeleteTransactionsOfType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	tType := chi.URLParam(r, "transactiontype")
	if !isValidTransactionType(tType) {
		respondWithError(w, 400, "Bad transaction type", nil)
		return
	}
	
	err := app.db.DeleteTransactionsOfType(ctx, database.DeleteTransactionsOfTypeParams{
		AccountID: acc.ID,
		TransactionType: tType,
	})
	if err != nil {
		respondWithError(w, 500, "Error deleting transactions from database", err)
		return
	}

	respondWithJSON(w, 200, "Transactions deleted successfully")
}

//Deletes all transaction records for a given account number
func (app *AppServer) handlerDeleteTransactionsForAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Delete transactions from database
	err := app.db.DeleteTransactionsForAccount(ctx, acc.ID)
	if err != nil {
		respondWithError(w, 500, "Error deleting transaction records from database", err)
		return
	}

	respondWithJSON(w, 200, "Transactions deleted successfully")
}

//Function deletes a transaction record from the database, based on transaction ID
func (app *AppServer) handlerDeleteTransactionRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Check if transaction exists
	_, err := app.db.GetSingleTransaction(ctx, acc.ID)
	if err != nil {
		respondWithError(w, 400, "No transaction record of that ID found", nil)
		return
	}

	//Delete transaction record from database
	err = app.db.DeleteTransaction(ctx, acc.ID)
	if err != nil {
		respondWithError(w, 500, "Error deleting transaction from database", err)
		return
	}

	respondWithJSON(w, 200, "Transaction deleted successfully")
}