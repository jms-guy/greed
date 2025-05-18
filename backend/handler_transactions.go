package main

import (
	"context"
	"encoding/json"
	"net/http"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/utils"
	"github.com/jms-guy/greed/models"
)

//Function will return all transactions for an account based on a category specified
func (cfg *apiConfig) handlerGetTransactionsOfCategory(w http.ResponseWriter, r *http.Request) {
	//Get account ID
	accId := r.PathValue("accountid")

	id, err := uuid.Parse(accId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Get category
	cat := r.PathValue("category")
	if cat == "" {
		respondWithError(w, 400, "Bad category given", nil)
		return
	}

	//Get transactions from database
	results, err := cfg.db.GetTransactionsOfCategory(context.Background(), database.GetTransactionsOfCategoryParams{
		AccountID: id,
		Category: cat,
	})
	if err != nil {
		respondWithError(w, 500, "Error retrieving transaction data", err)
		return
	}

	//Create return slice of structs
	transactions := []models.Transaction{}
	//Populate the return slice
	for _, result := range results {
		t := models.Transaction{
			ID: result.ID,
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
			Amount: result.Amount,
			Category: result.Category,
			Description: result.Description.String,
			TransactionDate: result.TransactionDate,
			TransactionType: result.TransactionType,
			CurrencyCode: result.CurrencyCode,
			AccountID: result.AccountID,
		}
		transactions = append(transactions, t)
	}

	respondWithJSON(w, 200, transactions)
}

//Function will return all transactions for an account based on given transaction type
func (cfg *apiConfig) handlerGetTransactionsofType(w http.ResponseWriter, r *http.Request) {
	//Get account ID
	accId := r.PathValue("accountid")

	id, err := uuid.Parse(accId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Get transaction type
	tType := r.PathValue("transactiontype")
	if !isValidTransactionType(tType) {
		respondWithError(w, 400, "Bad transaction type", nil)
		return
	}

	//Get transactions from database
	results, err := cfg.db.GetTransactionsOfType(context.Background(), database.GetTransactionsOfTypeParams{
		AccountID: id,
		TransactionType: tType,
	})
	if err != nil {
		respondWithError(w, 500, "Error retrieving transaction data", err)
		return
	}

	//Create return slice of structs
	transactions := []models.Transaction{}
	//Populate the return slice
	for _, result := range results {
		t := models.Transaction{
			ID: result.ID,
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
			Amount: result.Amount,
			Category: result.Category,
			Description: result.Description.String,
			TransactionDate: result.TransactionDate,
			TransactionType: result.TransactionType,
			CurrencyCode: result.CurrencyCode,
			AccountID: result.AccountID,
		}
		transactions = append(transactions, t)
	}

	respondWithJSON(w, 200, transactions)
}

//Function returns a single transaction from database, based on transaction ID
func (cfg *apiConfig) handlerGetTransactions(w http.ResponseWriter, r *http.Request) {
	//Get account ID
	accID := r.PathValue("accountid")

	id, err := uuid.Parse(accID)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Get transaction records
	results, err := cfg.db.GetTransactions(context.Background(), id)
	if err != nil {
		respondWithError(w, 500, "Error retrieving transaction data", err)
		return
	}

	//Create return slice of structs
	transactions := []models.Transaction{}
	//Populate the return slice
	for _, result := range results {
		t := models.Transaction{
			ID: result.ID,
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
			Amount: result.Amount,
			Category: result.Category,
			Description: result.Description.String,
			TransactionDate: result.TransactionDate,
			TransactionType: result.TransactionType,
			CurrencyCode: result.CurrencyCode,
			AccountID: result.AccountID,
		}
		transactions = append(transactions, t)
	}

	respondWithJSON(w, 200, transactions)
}

//Function will return a transaction result from database
func (cfg *apiConfig) handlerGetSingleTransaction(w http.ResponseWriter, r *http.Request) {
	//Get transaction ID
	transID := r.PathValue("transactionid")

	id, err := uuid.Parse(transID)
	if err != nil {
		respondWithError(w, 400, "Error parsing transaction ID", err)
		return
	}

	//Get transaction record from database
	result, err := cfg.db.GetSingleTransaction(context.Background(), id)
	if err != nil {
		respondWithError(w, 500, "Error retrieving transaction data", err)
		return
	}

	//Create response structure
	transaction := models.Transaction{
		ID: result.ID,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
		Amount: result.Amount,
		Category: result.Category,
		Description: result.Description.String,
		TransactionDate: result.TransactionDate,
		TransactionType: result.TransactionType,
		CurrencyCode: result.CurrencyCode,
		AccountID: result.AccountID,
	}

	respondWithJSON(w, 200, transaction)
}


//Function to create a transaction record in database
func (cfg *apiConfig) handlerCreateTransaction(w http.ResponseWriter, r *http.Request) {
	//Get account id
	accountId := r.PathValue("accountid")

	id, err := uuid.Parse(accountId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Decode request body
	decoder := json.NewDecoder(r.Body)
	params := models.CreateTransaction{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding request parameters", err)
		return
	}

	//Create database parameters struct
	sqlParams := database.CreateTransactionParams{
		ID: uuid.New(),
		Amount: params.Amount,
		Category: params.Category,
		Description: utils.CreateTextNullString(params.Description),
		TransactionDate: params.TransactionDate,
		TransactionType: params.TransactionType,
		CurrencyCode: params.CurrencyCode,
		AccountID: id,
	}

	//Create transaction in database
	transaction, err := cfg.db.CreateTransaction(context.Background(), sqlParams)
	if err != nil {
		respondWithError(w, 500, "Error creating transaction record in database", err)
		return
	}

	respondWithJSON(w, 201, transaction)
}