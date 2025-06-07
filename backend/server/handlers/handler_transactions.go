package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/utils"
	"github.com/jms-guy/greed/models"
)

//Function will return all transactions for an account based on a category specified
func (app *AppServer) HandlerGetTransactionsOfCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Get category
	cat := chi.URLParam(r, "category")
	if cat == "" {
		app.respondWithError(w, 400, "Bad category given", nil)
		return
	}

	//Get transactions from database
	results, err := app.Db.GetTransactionsOfCategory(ctx, database.GetTransactionsOfCategoryParams{
		AccountID: acc.ID,
		Category: cat,
	})
	if err != nil {
		app.respondWithError(w, 500, "Error retrieving transaction data", err)
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

	app.respondWithJSON(w, 200, transactions)
}

//Function will return all transactions for an account based on given transaction type
func (app *AppServer) HandlerGetTransactionsofType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Get transaction type
	tType := chi.URLParam(r, "transactiontype")
	/*if !isValidTransactionType(tType) {
		app.respondWithError(w, 400, "Bad transaction type", nil)
		return
	}*/

	//Get transactions from database
	results, err := app.Db.GetTransactionsOfType(ctx, database.GetTransactionsOfTypeParams{
		AccountID: acc.ID,
		TransactionType: tType,
	})
	if err != nil {
		app.respondWithError(w, 500, "Error retrieving transaction data", err)
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

	app.respondWithJSON(w, 200, transactions)
}

//Function returns a single transaction from database, based on transaction ID
func (app *AppServer) HandlerGetTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Get transaction records
	results, err := app.Db.GetTransactions(ctx, acc.ID)
	if err != nil {
		app.respondWithError(w, 500, "Error retrieving transaction data", err)
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

	app.respondWithJSON(w, 200, transactions)
}

//Function will return a transaction result from database
func (app *AppServer) HandlerGetSingleTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	transID := chi.URLParam(r, "transactionid")

	id, err := uuid.Parse(transID)
	if err != nil {
		app.respondWithError(w, 400, "Error parsing transaction ID", err)
		return
	}

	//Get transaction record from database
	result, err := app.Db.GetSingleTransaction(ctx, id)
	if err != nil {
		app.respondWithError(w, 500, "Error retrieving transaction data", err)
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

	app.respondWithJSON(w, 200, transaction)
}


//Function to create a transaction record in database
func (app *AppServer) HandlerCreateTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Decode request body
	decoder := json.NewDecoder(r.Body)
	params := models.CreateTransaction{}
	err := decoder.Decode(&params)
	if err != nil {
		app.respondWithError(w, 500, "Error decoding request parameters", err)
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
		AccountID: acc.ID,
	}

	//Create transaction in database
	transaction, err := app.Db.CreateTransaction(ctx, sqlParams)
	if err != nil {
		app.respondWithError(w, 500, "Error creating transaction record in database", err)
		return
	}

	app.respondWithJSON(w, 201, transaction)
}