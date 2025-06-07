package handlers

import (
	"net/http"
	"strconv"
	"github.com/go-chi/chi"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
)

func (app *AppServer) HandlerGetNetIncomeForMonth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Get month and year from path
	year := chi.URLParam(r, "year")
	y, err := strconv.Atoi(year)
	if err != nil {
        app.respondWithError(w, 400, "Invalid year format", nil)
		return
    }

	month := chi.URLParam(r, "month")
	m, err := strconv.Atoi(month)
	if err != nil || m < 1 || m > 12 {
        app.respondWithError(w, 400, "Invalid month format or out of range (1-12)", nil)
		return
    }

	//Get income amount from database calculation
	incAmount, err := app.Db.GetNetIncomeForMonth(ctx, database.GetNetIncomeForMonthParams{
		Year: int32(y),
		Month: int32(m),
		AccountID: acc.ID,
	})
	if err != nil {
		app.respondWithError(w, 500, "Error calculating income for month", err)
		return
	}

	//Make return struct
	netIncome := models.Income{
		Amount: incAmount,
		Year: y,
		Month: m,
	}

	app.respondWithJSON(w, 200, netIncome)
}

//Function returns the sum of all transactions with transaction_type of
//'debit' for a given account, to get expenses for a month
func (app *AppServer) HandlerGetExpensesForMonth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Get month and year from path
	year := chi.URLParam(r, "year")
	y, err := strconv.Atoi(year)
	if err != nil {
        app.respondWithError(w, 400, "Invalid year format", nil)
		return
    }

	month := chi.URLParam(r, "month")
	m, err := strconv.Atoi(month)
	if err != nil || m < 1 || m > 12 {
        app.respondWithError(w, 400, "Invalid month format or out of range (1-12)", nil)
		return
    }
    
    // Get expenses amount from database calculation
    expAmount, err := app.Db.GetExpensesForMonth(ctx, database.GetExpensesForMonthParams{
        Year: int32(y),
        Month: int32(m),
        AccountID: acc.ID,
    })
    if err != nil {
        app.respondWithError(w, 500, "Error calculating expenses for month", err)
        return
    }
    
    // Make return struct
    expenses := models.Expenses{  
        Amount: expAmount,
        Year: y,
        Month: m,
    }
    
    app.respondWithJSON(w, 200, expenses)
}

//Gets income for a given month from the database, by summing transactions
//of transaction_type 'credit' for a given account
func (app *AppServer) HandlerGetIncomeForMonth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Get month and year from path
	year := chi.URLParam(r, "year")
	y, err := strconv.Atoi(year)
	if err != nil {
        app.respondWithError(w, 400, "Invalid year format", nil)
		return
    }

	month := chi.URLParam(r, "month")
	m, err := strconv.Atoi(month)
	if err != nil || m < 1 || m > 12 {
        app.respondWithError(w, 400, "Invalid month format or out of range (1-12)", nil)
		return
    }
	//Get income amount from database calculation
	incAmount, err := app.Db.GetIncomeForMonth(ctx, database.GetIncomeForMonthParams{
		Year: int32(y),
		Month: int32(m),
		AccountID: acc.ID,
	})
	if err != nil {
		app.respondWithError(w, 500, "Error calculating income for month", err)
		return
	}

	//Make return struct
	income := models.Income{
		Amount: incAmount,
		Year: y,
		Month: m,
	}

	app.respondWithJSON(w, 200, income)
}