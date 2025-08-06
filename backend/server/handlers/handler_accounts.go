package handlers

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
	"net/http"
)

// Function will get all accounts for user
func (app *AppServer) HandlerGetAccountsForUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	//Get accounts for user from database
	accs, err := app.Db.GetAllAccountsForUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No accounts found for user", nil)
			return
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error retrieving accounts: %w", err))
		return
	}

	//Return slice of account structs
	var accounts []models.Account
	for _, account := range accs {
		result := models.Account{
			Id:               account.ID,
			CreatedAt:        account.CreatedAt,
			UpdatedAt:        account.UpdatedAt,
			Name:             account.Name,
			Type:             account.Type,
			Subtype:          account.Subtype.String,
			Mask:             account.Mask.String,
			OfficialName:     account.OfficialName.String,
			AvailableBalance: account.AvailableBalance.String,
			CurrentBalance:   account.CurrentBalance.String,
			IsoCurrencyCode:  account.IsoCurrencyCode.String,
		}
		accounts = append(accounts, result)
	}

	app.respondWithJSON(w, 200, accounts)
}

// Returns data of a single account, placed into context value through account middleware
func (app *AppServer) HandlerGetAccountData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	response := models.Account{
		Id:               acc.ID,
		CreatedAt:        acc.CreatedAt,
		UpdatedAt:        acc.UpdatedAt,
		Name:             acc.Name,
		Type:             acc.Type,
		Subtype:          acc.Subtype.String,
		Mask:             acc.Mask.String,
		OfficialName:     acc.OfficialName.String,
		AvailableBalance: acc.AvailableBalance.String,
		CurrentBalance:   acc.CurrentBalance.String,
		IsoCurrencyCode:  acc.IsoCurrencyCode.String,
		ItemId:           acc.ItemID,
	}

	app.respondWithJSON(w, 200, response)
}

// Function will delete an account from the database
func (app *AppServer) HandlerDeleteAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	accValue := ctx.Value(accountKey)
	acc, ok := accValue.(database.Account)
	if !ok {
		app.respondWithError(w, 400, "Bad account in context", nil)
		return
	}

	//Delete account from database based on user ID given
	err := app.Db.DeleteAccount(ctx, database.DeleteAccountParams{
		ID:     acc.ID,
		UserID: id,
	})
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting account: %w", err))
		return
	}

	app.respondWithJSON(w, 200, "Account deleted successfully")
}
