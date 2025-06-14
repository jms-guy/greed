package handlers

import (
	"fmt"
	"net/http"
	"github.com/jms-guy/greed/backend/internal/database"
)



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
