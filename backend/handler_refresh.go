package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/models"
)

//Handler function for generating a new JWT + refresh token for the user. Validates current
//refresh token, if invalidated then session delegation gets revoked
func (app *AppServer) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := models.RefreshRequest{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		app.respondWithError(w, 400, "Error decoding JSON parameters", err)
		return 
	}

	tokenHash := auth.HashRefreshToken(params.RefreshToken)

	token, err := app.db.GetToken(ctx, tokenHash)
	if err != nil {
		app.respondWithError(w, 401, "Refresh token not found", err)
		return 
	}

	if token.ExpiresAt.Before(time.Now()) {
		err  = app.db.ExpireToken(ctx, tokenHash)
		if err != nil {
			app.respondWithError(w, 500, "Error expiring refresh token", err)
			return
		}
		err = app.db.RevokeDelegationByID(ctx, token.DelegationID)
		if err != nil {
			app.respondWithError(w, 500, "Error revoking refresh session", err)
			return
		}
		app.respondWithError(w, 401, "Expired token", nil)
		return
	}
	 if token.IsUsed {
		err = app.db.RevokeDelegationByID(ctx, token.DelegationID)
		if err != nil {
			app.respondWithError(w, 500, "Error revoking refresh session", err)
			return
		}
		err = app.db.ExpireAllDelegationTokens(ctx, token.DelegationID)
		if err != nil {
			app.respondWithError(w, 500, "Error expiring tokens of delegation", err)
		}

		app.respondWithError(w, 401, "Refresh token has already been used", nil)
		return
	}

	newJWT, err := auth.MakeJWT(app.config, token.UserID)
	if err != nil {
		app.respondWithError(w, 500, "Error creating JWT", err)
		return 
	}

	del, err := app.db.GetDelegation(ctx, token.DelegationID)
	if err != nil {
		app.respondWithError(w, 401, "Session delegation not found", err)
		return
	}

	newToken, err := auth.MakeRefreshToken(app.db, token.UserID, del)
	if err != nil {
		app.respondWithError(w, 500, "Error creating new refresh token", err)
		return
	}

	err = app.db.ExpireToken(ctx, tokenHash)
	if err != nil {
		app.respondWithError(w, 500, "Error expiring refresh token", err)
		return
	}

	response := models.RefreshResponse{
		RefreshToken: 	newToken,
		AccessToken: 	newJWT,
		TokenType: 		"Bearer",
	}

	app.respondWithJSON(w, 200, response)
}