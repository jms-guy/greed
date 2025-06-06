package main

import (
	"encoding/json"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/utils"
	"github.com/jms-guy/greed/models"
)

//Function will update the category of a transaction record in database
func (app *AppServer) handlerUpdateTransactionCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	transID := chi.URLParam(r, "transactionid")

	id, err := uuid.Parse(transID)
	if err != nil {
		app.respondWithError(w, 400, "Error parsing transaction ID", err)
		return
	}

	//Decode request body
	decoder := json.NewDecoder(r.Body)
	params := models.UpdateCategory{}
	err = decoder.Decode(&params)
	if err != nil {
		app.respondWithError(w, 500, "Error decoding request parameters", err)
		return
	}

	//Update transaction category in database
	result, err := app.db.UpdateTransactionCategory(ctx, database.UpdateTransactionCategoryParams{
		Category: params.Category,
		ID: id,
	})
	if err != nil {
		app.respondWithError(w, 500, "Error updating transaction category in database", err)
		return
	}

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

//Function to update the description of a transaction based on transaction ID
func (app *AppServer) handlerUpdateTransactionDescription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	transID := chi.URLParam(r, "transactionid")

	id, err := uuid.Parse(transID)
	if err != nil {
		app.respondWithError(w, 400, "Error parsing transaction ID", err)
		return
	}

	//Decode request body
	decoder := json.NewDecoder(r.Body)
	params := models.UpdateDescription{}
	err = decoder.Decode(&params)
	if err != nil {
		app.respondWithError(w, 500, "Error decoding request parameters", err)
		return
	}

	//Update transaction record in database
	result, err := app.db.UpdateTransactionDescription(ctx, database.UpdateTransactionDescriptionParams{
		Description: utils.CreateTextNullString(params.Description),
		ID: id,
	})
	if err != nil {
		app.respondWithError(w, 500, "Error updating transaction record in database", err)
		return
	}

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
