package handlers

import (
	"fmt"
	"net/http"

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

