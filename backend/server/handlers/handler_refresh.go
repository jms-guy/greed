package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jms-guy/greed/models"
)

// Handler function for generating a new JWT + refresh token for the user. Validates current
// refresh token, if invalidated then session delegation gets revoked
func (app *AppServer) HandlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := models.RefreshRequest{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		app.respondWithError(w, 400, "Error decoding JSON parameters", err)
		return
	}

	tokenHash := app.Auth.HashRefreshToken(params.RefreshToken)

	token, err := app.Db.GetToken(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 401, "Refresh token not found", err)
			return
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting refresh token: %w", err))
	}

	if token.ExpiresAt.Before(time.Now()) {
		err = app.TxnUpdater.ExpireDelegation(ctx, tokenHash, token)
		if err != nil {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error expiring delegation token: %w", err))
			return
		}
		app.respondWithError(w, 401, "Token is expired", nil)
		return
	}
	if token.IsUsed {
		err = app.TxnUpdater.RevokeDelegation(ctx, token)
		if err != nil {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error revoking delegation: %w", err))
			return
		}
		app.respondWithError(w, 401, "Refresh token has already been used", nil)
		return
	}

	newJWT, err := app.Auth.MakeJWT(app.Config, token.UserID)
	if err != nil {
		app.respondWithError(w, 500, "Error creating JWT", err)
		return
	}

	del, err := app.Db.GetDelegation(ctx, token.DelegationID)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 401, "Session delegation not found", err)
			return
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting delegation: %w", err))
		return
	}

	newToken, err := app.Auth.MakeRefreshToken(app.Db, token.UserID, del)
	if err != nil {
		app.respondWithError(w, 500, "Error creating new refresh token", err)
		return
	}

	err = app.Db.ExpireToken(ctx, tokenHash)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error expiring refresh token: %w", err))
		return
	}

	response := models.RefreshResponse{
		RefreshToken: newToken,
		AccessToken:  newJWT,
		TokenType:    "Bearer",
	}

	app.respondWithJSON(w, 200, response)
}
