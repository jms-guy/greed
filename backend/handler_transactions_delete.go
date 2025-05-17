package main

import (
	"context"
	"net/http"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/internal/database"
)

//Function deletes all transaction records for an account of a given category
func (cfg *apiConfig) handlerDeleteTransactionsOfCategory(w http.ResponseWriter, r *http.Request) {
	//Get account ID
	accId := r.PathValue("accountid")

	id, err := uuid.Parse(accId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Get category (user flexibility here)
	cat := r.PathValue("category")
	
	if cat == "" {
		respondWithError(w, 400, "Bad category given", nil)
		return
	}

	//Delete transactions from database
	err = cfg.db.DeleteTransactionsOfCategory(context.Background(), database.DeleteTransactionsOfCategoryParams{
		AccountID: id,
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
func (cfg *apiConfig) handlerDeleteTransactionsOfType(w http.ResponseWriter, r *http.Request) {
	accId := r.PathValue("accountid")

	id, err := uuid.Parse(accId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	tType := r.PathValue("transactiontype")
	if !isValidTransactionType(tType) {
		respondWithError(w, 400, "Bad transaction type", nil)
		return
	}
	
	err = cfg.db.DeleteTransactionsOfType(context.Background(), database.DeleteTransactionsOfTypeParams{
		AccountID: id,
		TransactionType: tType,
	})
	if err != nil {
		respondWithError(w, 500, "Error deleting transactions from database", err)
		return
	}

	respondWithJSON(w, 200, "Transactions deleted successfully")
}

//Deletes all transaction records for a given account number
func (cfg *apiConfig) handlerDeleteTransactionsForAccount(w http.ResponseWriter, r *http.Request) {
	//Get account ID
	accId := r.PathValue("accountid")

	id, err := uuid.Parse(accId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Delete transactions from database
	err = cfg.db.DeleteTransactionsForAccount(context.Background(), id)
	if err != nil {
		respondWithError(w, 500, "Error deleting transaction records from database", err)
		return
	}

	respondWithJSON(w, 200, "Transactions deleted successfully")
}

//Function deletes a transaction record from the database, based on transaction ID
func (cfg *apiConfig) handlerDeleteTransactionRecord(w http.ResponseWriter, r *http.Request) {
	//Get transaction ID
	transID := r.PathValue("transactionid")

	id, err := uuid.Parse(transID)
	if err != nil {
		respondWithError(w, 400, "Error parsing transaction ID", err)
		return
	}

	//Delete transaction record from database
	err = cfg.db.DeleteTransaction(context.Background(), id)
	if err != nil {
		respondWithError(w, 500, "Error deleting transaction from database", err)
		return
	}

	respondWithJSON(w, 200, "Transaction deleted successfully")
}