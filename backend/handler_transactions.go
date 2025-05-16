package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/internal/database"
	"github.com/jms-guy/greed/internal/utils"
)

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
	transactions := []Transaction{}
	//Populate the return slice
	for _, result := range results {
		t := Transaction{
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
	transaction := Transaction{
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

//Function will update the category of a transaction record in database
func (cfg *apiConfig) handlerUpdateTransactionCategory(w http.ResponseWriter, r *http.Request) {
	//Parameters needed from request
	type parameters struct {
		Category		string `json:"category"`
	}

	//Get transaction ID
	transID := r.PathValue("transactionid")

	id, err := uuid.Parse(transID)
	if err != nil {
		respondWithError(w, 400, "Error parsing transaction ID", err)
		return
	}

	//Decode request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding request parameters", err)
		return
	}

	//Update transaction category in database
	err = cfg.db.UpdateTransactionCategory(context.Background(), database.UpdateTransactionCategoryParams{
		Category: params.Category,
		ID: id,
	})
	if err != nil {
		respondWithError(w, 500, "Error updating transaction category in database", err)
		return
	}

	respondWithJSON(w, 200, Transaction{Category: params.Category})
}

//Function to update the description of a transaction based on transaction ID
func (cfg *apiConfig) handlerUpdateTransactionDescription(w http.ResponseWriter, r *http.Request) {
	//Parameters needed from request
	type parameters struct {
		Description		string `json:"description"`
	}

	//Get transaction ID
	transID := r.PathValue("transactionid")

	id, err := uuid.Parse(transID)
	if err != nil {
		respondWithError(w, 400, "Error parsing transaction ID", err)
		return
	}

	//Decode request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding request parameters", err)
		return
	}

	//Update transaction record in database
	err = cfg.db.UpdateTransactionDescription(context.Background(), database.UpdateTransactionDescriptionParams{
		Description: utils.CreateTextNullString(params.Description),
		ID: id,
	})
	if err != nil {
		respondWithError(w, 500, "Error updating transaction record in database", err)
		return
	}

	respondWithJSON(w, 200, Transaction{Description: params.Description})
}

//Function to create a transaction record in database
func (cfg *apiConfig) handlerCreateTransaction(w http.ResponseWriter, r *http.Request) {
	//Parameters needed from request
	type parameters struct {
		Amount			string 	  `json:"amount"`
		Category		string	  `json:"category"`
		Description		string	  `json:"description"`
		TransactionDate time.Time `json:"transaction_date"`
		TransactionType string	  `json:"transaction_type"`
		CurrencyCode	string	  `json:"currency_code"`
		AccountID		uuid.UUID `json:"account_id"`
	}

	//Get account id
	accountId := r.PathValue("accountId")

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