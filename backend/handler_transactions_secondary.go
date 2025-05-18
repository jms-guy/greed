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


//Function will update the category of a transaction record in database
func (cfg *apiConfig) handlerUpdateTransactionCategory(w http.ResponseWriter, r *http.Request) {
	//Get transaction ID
	transID := r.PathValue("transactionid")

	id, err := uuid.Parse(transID)
	if err != nil {
		respondWithError(w, 400, "Error parsing transaction ID", err)
		return
	}

	//Decode request body
	decoder := json.NewDecoder(r.Body)
	params := models.UpdateCategory{}
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

	respondWithJSON(w, 200, models.Transaction{Category: params.Category})
}

//Function to update the description of a transaction based on transaction ID
func (cfg *apiConfig) handlerUpdateTransactionDescription(w http.ResponseWriter, r *http.Request) {
	//Get transaction ID
	transID := r.PathValue("transactionid")

	id, err := uuid.Parse(transID)
	if err != nil {
		respondWithError(w, 400, "Error parsing transaction ID", err)
		return
	}

	//Decode request body
	decoder := json.NewDecoder(r.Body)
	params := models.UpdateDescription{}
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

	respondWithJSON(w, 200, models.Transaction{Description: params.Description})
}
