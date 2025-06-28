package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"github.com/go-chi/chi/v5"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/utils"
	"github.com/jms-guy/greed/models"
)

//Initializes a map, the key being expected query parameters, the value being the expected data type it should be.
//All parameters are optional, but "date" is incompatible with ("start", "end")
func makeQueryRules() map[string]string {
	rules := map[string]string{
		"merchant":		"string",
		"category":		"string",
		"channel":		"string",
		"date":			"date",
		"start": 		"date",
		"end": 			"date",
		"min":			"number",
		"max":			"number",
		"limit":		"number",
		"order":		"string",
	}
	return rules
}

//Handler gets transaction records for an account, parsing query parameters to deliver correct records
func (app *AppServer) HandlerGetTransactionsForAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account data in context", nil)
		return 
	}

	params := r.URL.Query()

	//If more than one value present for a single query key, only the first value will be taken
	qv := utils.NewQueryValidator()
	queries, errs := qv.ValidateQuery(params, makeQueryRules())
	if len(errs) != 0 {
		app.respondWithError(w, 400, fmt.Sprintf("Bad query parameter: %v", errs), nil)
		return 
	}

	dbQuery, args, err := utils.BuildSqlQuery(queries, acc.ID)
	if err != nil {
		app.respondWithError(w, 500, fmt.Sprintf("Error building database query: %s", err), err)
		return
	}

	rows, err := app.Database.QueryContext(ctx, dbQuery, args...)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error executing query: %w", err))
		return 
	}
	defer rows.Close()

	var txns []database.Transaction
	for rows.Next() {
		var t database.Transaction
		if err := rows.Scan(
			&t.ID,
			&t.AccountID,
			&t.Amount,
			&t.IsoCurrencyCode,
			&t.Date,
			&t.MerchantName,
			&t.PaymentChannel,
			&t.PersonalFinanceCategory,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			app.respondWithError(w, 500, "Error parsing transaction rows", err)
			return 
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		app.respondWithError(w, 500, "Error encounted during query", err)
		return 
	}
	if err := rows.Close(); err != nil {
		app.respondWithError(w, 500, "Error closing rows", err)
		return 
	}

	var transactions []models.Transaction
	for _, t := range txns {
		respond := models.Transaction{
			Id: t.ID,
			AccountId: t.AccountID,
			Amount: t.Amount,
			IsoCurrencyCode: t.IsoCurrencyCode.String,
			Date: t.Date.Time,
			MerchantName: t.MerchantName.String,
			PaymentChannel: t.PaymentChannel,
			PersonalFinanceCategory: t.PersonalFinanceCategory,
		}

		transactions = append(transactions, respond)
	}
	
	app.respondWithJSON(w, 200, transactions)
}

//Gets income, net income, and expenses for a given month from the database, by summing transactional records 
func (app *AppServer) HandlerGetMonetaryDataForMonth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	if acc.Type != "depository" && acc.Type != "credit" {
		app.respondWithError(w, 400, "Endpoint only works with debit or credit accounts", nil)
		return
	}

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

	incAmount, err := app.Db.GetDataForMonth(ctx, database.GetDataForMonthParams{
		Year: int32(y),
		Month: int32(m),
		AccountID: acc.ID,
	})
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error calculating income: %w", err))
		return
	}

	income := models.MonetaryData{
		Income: incAmount.Income,
		Expenses: incAmount.Expenses,
		NetIncome: incAmount.NetIncome,
		Year: y,
		Month: m,
	}

	app.respondWithJSON(w, 200, income)
}

//Gets historical income/expense data for account. Aggregates transaction amounts, and sorts based on month
func (app *AppServer) HandlerGetMonetaryData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	if acc.Type != "depository" && acc.Type != "credit" {
		app.respondWithError(w, 400, "Endpoint only works with debit or credit accounts", nil)
		return
	}

	data, err := app.Db.GetDataForAllMonths(ctx, acc.ID)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting income/expense data: %w", err))
		return 
	}

	var response []models.MonetaryData
	for _, record := range data {

		incData := models.MonetaryData{
			Income: record.Income,
			Expenses: record.Expenses,
			NetIncome: record.NetIncome,
			Year: int(record.Year),
			Month: int(record.Month),
		}

		response = append(response, incData)
	}

	app.respondWithJSON(w, 200, response)
}

//Deletes all transaction records for a given account ID
func (app *AppServer) HandlerDeleteTransactionsForAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	err := app.Db.DeleteTransactionsForAccount(ctx, acc.ID)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting transaction records: %w", err))
		return
	}

	app.respondWithJSON(w, 200, "Transactions deleted successfully")
}