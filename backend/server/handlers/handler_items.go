package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
)

//Grabs item records for a user from database, returning names and item IDs
func (app *AppServer) HandlerGetItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	items, err := app.Db.GetItemsByUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No items found for user", nil)
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting item records: %w", err))
		return
	}
	response := struct {
		Items []models.ItemName `json:"items"`
	}{}
	
	for _, item := range items {
		nickname := ""
		if item.Nickname.Valid {
			nickname = item.Nickname.String
		}
		response.Items = append(response.Items, models.ItemName{
			ItemId: item.ID,
			Nickname: nickname,
			InstitutionName: item.InstitutionName,
		})
	}
	app.respondWithJSON(w, 200, response)
}

//Updates item's nickname field in database
func (app *AppServer) HandlerUpdateItemName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	request := models.UpdateItemName{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		app.respondWithError(w, 400, "Bad request data", err)
		return 
	}

	itemID := chi.URLParam(r, "item-id")

	params := database.UpdateNicknameParams{
		Nickname: sql.NullString{String: request.Nickname, Valid: true},
		ID: itemID,
		UserID: id,
	}

	err := app.Db.UpdateNickname(ctx, params)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error updating item name: %w", err))
		return 
	}

	app.respondWithJSON(w, 200, "Item name updated successfully")
}


//Endpoint deletes a user's item from database
func (app *AppServer) HandlerDeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	tokenValue := ctx.Value(accessTokenKey)
	accessToken, ok := tokenValue.(string)
	if !ok {
		app.respondWithError(w, 400, "Bad access token in context", nil)
		return 
	}

	itemID := chi.URLParam(r, "item-id")

	params := database.DeleteItemParams{
		ID: itemID,
		UserID: id,
	}

	err := app.Db.DeleteItem(ctx, params)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting item record: %w", err))
		return 
	}

	err = app.PService.RemoveItem(ctx, accessToken)
	if err != nil {
		app.respondWithError(w, 500, "Service error", fmt.Errorf("error removing item from plaid databases: %w", err))
	}

	app.respondWithJSON(w, 200, "Item deleted successfully")
}

//Gets accounts only for a user's specific item
func (app *AppServer) HandlerGetAccountsForItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	itemID := chi.URLParam(r, "item-id")

	accs, err := app.Db.GetAccountsForItem(ctx, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No accounts found for item", nil)
			return 
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting accounts for item: %w", err))
		return 
	}

	//Return slice of account structs
	var accounts []models.Account
	for _, account := range accs {
		result := models.Account{
			Id: account.ID,
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
			Name: account.Name,
			Type: account.Type,
			Subtype: account.Subtype.String,
			Mask: account.Mask.String,
			OfficialName: account.OfficialName.String,
			AvailableBalance: account.AvailableBalance.String,
			CurrentBalance: account.CurrentBalance.String,
			IsoCurrencyCode: account.IsoCurrencyCode.String,
		}
		accounts = append(accounts, result)
	}
	
	app.respondWithJSON(w, 200, accounts)
}

//Searches database for records sent by Plaid webhook related to user's items
func (app *AppServer) HandlerGetWebhookRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok || id == uuid.Nil {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	records, err := app.Db.GetWebhookRecords(ctx, id) 
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 404, "No webhook records found", nil) 
			return
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting webhook records for user: %w", err))
		return
	}

	var webhookRecords []models.WebhookRecord
	for _, record := range records { 
		foundRecord := models.WebhookRecord{
			WebhookType: record.WebhookType,
			WebhookCode: record.WebhookCode,
			UserID: id,
			ItemID: record.ItemID,
			CreatedAt: record.CreatedAt.Format("2006-01-02"),
		}
		webhookRecords = append(webhookRecords, foundRecord)
	}

	app.respondWithJSON(w, 200, webhookRecords)
}