package handlers

import (
	"net/http"
)

func (app *AppServer) HandlerGetTransactionsForAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(string)
	if !ok {
		app.respondWithError(w, 400, "Bad account data in context", nil)
		return 
	}

	
}

