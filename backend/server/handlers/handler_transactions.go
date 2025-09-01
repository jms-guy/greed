package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
)

// Initializes a map, the key being expected query parameters, the value being the expected data type it should be.
// All parameters are optional, but "date" is incompatible with ("start", "end")
func makeQueryRules() map[string]string {
	rules := map[string]string{
		"merchant": "string",
		"category": "string",
		"channel":  "string",
		"date":     "date",
		"start":    "date",
		"end":      "date",
		"min":      "number",
		"max":      "number",
		"limit":    "number",
		"summary":  "string",
	}
	return rules
}

// Handler gets transaction records for an account, parsing query parameters to deliver correct records
func (app *AppServer) HandlerGetTransactionsForAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account data in context", nil)
		return
	}

	params := r.URL.Query()

	// If more than one value present for a single query key, only the first value will be taken
	queries, errs := app.Querier.ValidateQuery(params, makeQueryRules())
	if len(errs) != 0 {
		app.respondWithError(w, 400, fmt.Sprintf("Bad query parameter: %v", errs), nil)
		return
	}

	// If summary flag was used
	date, ok := queries["date"] // Date being a string in format ("2006-01-02")
	if queries["summary"] == "true" && ok {
		vals := strings.Split(date, "-")
		year, month := vals[0], vals[1]

		yearVal, err := strconv.Atoi(year)
		if err != nil {
			app.respondWithError(w, 500, fmt.Sprintf("Error converting date value: %s", err), err)
			return
		}
		monthVal, err := strconv.Atoi(month)
		if err != nil {
			app.respondWithError(w, 500, fmt.Sprintf("Error converting date value: %s", err), err)
			return
		}

		params := database.GetMerchantSummaryByMonthParams{
			AccountID: acc.ID,
			// #nosec G115 G109 - int32 is fine for these values
			Year: int32(yearVal),
			// #nosec G115 G109
			Month: int32(monthVal),
		}
		summaries, err := app.Db.GetMerchantSummaryByMonth(ctx, params)
		if err != nil {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error executing query: %w", err))
			return
		}

		var responseSummary []models.MerchantSummary
		for _, sum := range summaries {
			total := strconv.FormatFloat(sum.TotalAmount, 'f', 2, 64)
			s := models.MerchantSummary{
				Merchant:    sum.Merchant.String,
				TxnCount:    sum.TxnCount,
				Category:    sum.Category,
				TotalAmount: total,
				Month:       sum.Month,
			}
			responseSummary = append(responseSummary, s)
		}
		app.respondWithJSON(w, 200, responseSummary)
		return
	} else if queries["summary"] == "true" { // Summary flag, but no date flag
		summary, err := app.Db.GetMerchantSummary(ctx, acc.ID)
		if err != nil {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error executing query: %w", err))
			return
		}

		var responseSummary []models.MerchantSummary
		for _, sum := range summary {
			total := strconv.FormatFloat(sum.TotalAmount, 'f', 2, 64)
			s := models.MerchantSummary{
				Merchant:    sum.Merchant.String,
				TxnCount:    sum.TxnCount,
				Category:    sum.Category,
				TotalAmount: total,
				Month:       sum.Month,
			}
			responseSummary = append(responseSummary, s)
		}
		app.respondWithJSON(w, 200, responseSummary)
		return
	}

	// No summary flag, continue with query
	dbQuery, args, err := app.Querier.BuildSqlQuery(queries, acc.ID)
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
		app.respondWithError(w, 500, "Error encountered during query", err)
		return
	}
	if err := rows.Close(); err != nil {
		app.respondWithError(w, 500, "Error closing rows", err)
		return
	}

	var transactions []models.Transaction
	for _, t := range txns {
		respond := models.Transaction{
			Id:                      t.ID,
			AccountId:               t.AccountID,
			Amount:                  t.Amount,
			IsoCurrencyCode:         t.IsoCurrencyCode.String,
			Date:                    t.Date.Time,
			MerchantName:            t.MerchantName.String,
			PaymentChannel:          t.PaymentChannel,
			PersonalFinanceCategory: t.PersonalFinanceCategory,
		}

		transactions = append(transactions, respond)
	}

	app.respondWithJSON(w, 200, transactions)
}

// Gets income, net income, and expenses for a given month from the database, by summing transactional records
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

	incAmount, err := app.Db.GetMonetaryDataForMonth(ctx, database.GetMonetaryDataForMonthParams{
		// #nosec G115 G109 - int32 is fine for these values
		Year: int32(y),
		// #nosec G115 G109
		Month:     int32(m),
		AccountID: acc.ID,
	})
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error calculating income: %w", err))
		return
	}

	date := fmt.Sprintf("%s-%s", strconv.Itoa(int(y)), strconv.Itoa(int(m)))

	income := models.MonetaryData{
		Income:    incAmount.Income,
		Expenses:  incAmount.Expenses,
		NetIncome: incAmount.NetIncome,
		Date:      date,
	}

	app.respondWithJSON(w, 200, income)
}

// Gets historical income/expense data for account. Aggregates transaction amounts, and sorts based on month
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

	data, err := app.Db.GetMonetaryDataForAllMonths(ctx, acc.ID)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting income/expense data: %w", err))
		return
	}

	var response []models.MonetaryData
	for _, record := range data {
		date := fmt.Sprintf("%s-%s", strconv.Itoa(int(record.Year)), strconv.Itoa(int(record.Month)))

		incData := models.MonetaryData{
			Income:    record.Income,
			Expenses:  record.Expenses,
			NetIncome: record.NetIncome,
			Date:      date,
		}

		response = append(response, incData)
	}

	app.respondWithJSON(w, 200, response)
}

// Deletes all transaction records for a given account ID
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

// Handler gets recurring transaction data for a single account
func (app *AppServer) HandlerGetRecurringData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	var recurringStreams []models.RecurringStream
	var streamConnections []models.TransactionsToStream
	var summary models.StreamSummary

	streams, err := app.Db.GetStreamsForAcc(ctx, acc.ID) // Get recurring streams for account
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting recurring stream data: %w", err))
		return
	}

	for _, stream := range streams {
		newStream := models.RecurringStream{
			ID:                stream.ID,
			AccountID:         stream.AccountID,
			Description:       stream.Description,
			MerchantName:      stream.MerchantName.String,
			Frequency:         stream.Frequency,
			IsActive:          stream.IsActive,
			PredictedNextDate: stream.PredictedNextDate.String,
			StreamType:        stream.StreamType,
		}

		recurringStreams = append(recurringStreams, newStream) // Append each stream into return struct

		summary.TotalStreams++
		if stream.IsActive {
			summary.ActiveStreams++
		} else {
			summary.InactiveStreams++
		}

		connections, err := app.Db.GetTransactionsToStreamConnections(ctx, stream.ID) // Get connections for each stream
		if err != nil {                                                               // Strict error check, loosen in future
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting transaction-to-stream record: %w", err))
			return
		}

		for _, c := range connections {
			newConnection := models.TransactionsToStream{
				TransactionID: c.TransactionID,
				StreamID:      c.StreamID,
			}

			streamConnections = append(streamConnections, newConnection) // Append the connections into return struct
		}
	}

	recurringData := models.RecurringData{
		Streams:     recurringStreams,
		Connections: streamConnections,
		Summary:     summary,
	}

	app.respondWithJSON(w, 200, recurringData)
}
