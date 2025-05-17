package main

import (
	"context"
	"net/http"
	"strconv"
	"time"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/internal/database"
)

func (cfg *apiConfig) handlerGetNetIncomeForMonth(w http.ResponseWriter, r *http.Request) {
	//Get account ID
	accId := r.PathValue("accountid")

	id, err := uuid.Parse(accId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Get month and year from path
	year := r.PathValue("year")
	y, err := strconv.Atoi(year)
	if err != nil {
        // Default to current year if not specified or invalid
        y = time.Now().Year()
    }

	month := r.PathValue("month")
	m, err := strconv.Atoi(month)
	if err != nil || m < 1 || m > 12 {
        // Default to current month if not specified or invalid
        m = int(time.Now().Month())
    }

	//Get income amount from database calculation
	incAmount, err := cfg.db.GetNetIncomeForMonth(context.Background(), database.GetNetIncomeForMonthParams{
		Year: int32(y),
		Month: int32(m),
		AccountID: id,
	})
	if err != nil {
		respondWithError(w, 500, "Error calculating income for month", err)
		return
	}

	//Make return struct
	netIncome := Income{
		Amount: incAmount,
		Year: y,
		Month: m,
	}

	respondWithJSON(w, 200, netIncome)
}

//Function returns the sum of all transactions with transaction_type of
//'debit' for a given account, to get expenses for a month
func (cfg *apiConfig) handlerGetExpensesForMonth(w http.ResponseWriter, r *http.Request) {
    // Get account ID
    accId := r.PathValue("accountid")
    
    id, err := uuid.Parse(accId)
    if err != nil {
        respondWithError(w, 400, "Error parsing account ID", err)
        return
    }

	//Get month and year from path
	year := r.PathValue("year")
	y, err := strconv.Atoi(year)
	if err != nil {
        // Default to current year if not specified or invalid
        y = time.Now().Year()
    }

	month := r.PathValue("month")
	m, err := strconv.Atoi(month)
	if err != nil || m < 1 || m > 12 {
        // Default to current month if not specified or invalid
        m = int(time.Now().Month())
    }
    
    // Get expenses amount from database calculation
    expAmount, err := cfg.db.GetExpensesForMonth(context.Background(), database.GetExpensesForMonthParams{
        Year: int32(y),
        Month: int32(m),
        AccountID: id,
    })
    if err != nil {
        respondWithError(w, 500, "Error calculating expenses for month", err)
        return
    }
    
    // Make return struct
    expenses := Expenses{  
        Amount: expAmount,
        Year: y,
        Month: m,
    }
    
    respondWithJSON(w, 200, expenses)
}

//Gets income for a given month from the database, by summing transactions
//of transaction_type 'credit' for a given account
func (cfg *apiConfig) handlerGetIncomeForMonth(w http.ResponseWriter, r *http.Request) {
	//Get account ID
	accId := r.PathValue("accountid")

	id, err := uuid.Parse(accId)
	if err != nil {
		respondWithError(w, 400, "Error parsing account ID", err)
		return
	}

	//Get month and year from path
	year := r.PathValue("year")
	y, err := strconv.Atoi(year)
	if err != nil {
        // Default to current year if not specified or invalid
        y = time.Now().Year()
    }

	month := r.PathValue("month")
	m, err := strconv.Atoi(month)
	if err != nil || m < 1 || m > 12 {
        // Default to current month if not specified or invalid
        m = int(time.Now().Month())
    }

	//Get income amount from database calculation
	incAmount, err := cfg.db.GetIncomeForMonth(context.Background(), database.GetIncomeForMonthParams{
		Year: int32(y),
		Month: int32(m),
		AccountID: id,
	})
	if err != nil {
		respondWithError(w, 500, "Error calculating income for month", err)
		return
	}

	//Make return struct
	income := Income{
		Amount: incAmount,
		Year: y,
		Month: m,
	}

	respondWithJSON(w, 200, income)
}