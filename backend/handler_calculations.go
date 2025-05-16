package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/internal/database"
)

func (cfg *apiConfig) handlerGetNetIncomeForMonth(w http.ResponseWriter, r *http.Request) {
	//Parameters needed in request
	type parameters struct {
		Year		int `json:"year"`
		Month		int `json:"month"`
	}

	//Get account ID
	accId := r.PathValue("accountid")

	id, err := uuid.Parse(accId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Decode request body
	params := parameters{}
	if err = json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, 500, "Error decoding request body", err)
		return
	}

	//Make sure date values are present
	if (params.Month == 0) || (params.Year == 0) {
		respondWithError(w, 400, "Invalid date", nil)
		return
	}

	//Get income amount from database calculation
	incAmount, err := cfg.db.GetNetIncomeForMonth(context.Background(), database.GetNetIncomeForMonthParams{
		Year: int32(params.Year),
		Month: int32(params.Month),
		AccountID: id,
	})
	if err != nil {
		respondWithError(w, 500, "Error calculating income for month", err)
		return
	}

	//Make return struct
	netIncome := Income{
		Amount: incAmount,
		Year: params.Year,
		Month: params.Month,
	}

	respondWithJSON(w, 200, netIncome)
}

//Function returns the sum of all transactions with transaction_type of
//'debit' for a given account, to get expenses for a month
func (cfg *apiConfig) handlerGetExpensesForMonth(w http.ResponseWriter, r *http.Request) {
    //Parameters needed in request
    type parameters struct {
        Year   int `json:"year"`
        Month  int `json:"month"`
    }
    
    // Get account ID
    accId := r.PathValue("accountid")
    
    id, err := uuid.Parse(accId)
    if err != nil {
        respondWithError(w, 400, "Error parsing account ID", err)
        return
    }
    
    // Decode request body
    params := parameters{}
    if err = json.NewDecoder(r.Body).Decode(&params); err != nil {
        respondWithError(w, 500, "Error decoding request body", err)
        return
    }
    
    // Make sure date values are present
    if (params.Month == 0) || (params.Year == 0) {
        respondWithError(w, 400, "Invalid date", nil)
        return
    }
    
    // Get expenses amount from database calculation
    expAmount, err := cfg.db.GetExpensesForMonth(context.Background(), database.GetExpensesForMonthParams{
        Year: int32(params.Year),
        Month: int32(params.Month),
        AccountID: id,
    })
    if err != nil {
        respondWithError(w, 500, "Error calculating expenses for month", err)
        return
    }
    
    // Make return struct
    expenses := Income{  // You might want to use a different struct name
        Amount: expAmount,
        Year: params.Year,
        Month: params.Month,
    }
    
    respondWithJSON(w, 200, expenses)
}

//Gets income for a given month from the database, by summing transactions
//of transaction_type 'credit' for a given account
func (cfg *apiConfig) handlerGetIncomeForMonth(w http.ResponseWriter, r *http.Request) {
	//Parameters needed in request
	type parameters struct {
		Year		int `json:"year"`
		Month		int `json:"month"`
	}

	//Get account ID
	accId := r.PathValue("accountid")

	id, err := uuid.Parse(accId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Decode request body
	params := parameters{}
	if err = json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, 500, "Error decoding request body", err)
		return
	}

	//Make sure date values are present
	if (params.Month == 0) || (params.Year == 0) {
		respondWithError(w, 400, "Invalid date", nil)
		return
	}

	//Get income amount from database calculation
	incAmount, err := cfg.db.GetIncomeForMonth(context.Background(), database.GetIncomeForMonthParams{
		Year: int32(params.Year),
		Month: int32(params.Month),
		AccountID: id,
	})
	if err != nil {
		respondWithError(w, 500, "Error calculating income for month", err)
		return
	}

	//Make return struct
	income := Income{
		Amount: incAmount,
		Year: params.Year,
		Month: params.Month,
	}

	respondWithJSON(w, 200, income)
}