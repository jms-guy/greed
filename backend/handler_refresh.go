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
func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := models.RefreshRequest{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, 400, "Error decoding JSON parameters", err)
		return 
	}

	tokenHash := auth.HashRefreshToken(params.RefreshToken)

	token, err := cfg.db.GetToken(ctx, tokenHash)
	if err != nil {
		respondWithError(w, 401, "Refresh token not found", err)
		return 
	}

	if token.ExpiresAt.Before(time.Now()) {
		err  = cfg.db.ExpireToken(ctx, tokenHash)
		if err != nil {
			respondWithError(w, 500, "Error expiring refresh token", err)
			return
		}
		err = cfg.db.RevokeDelegationByID(ctx, token.DelegationID)
		if err != nil {
			respondWithError(w, 500, "Error revoking refresh session", err)
			return
		}
		respondWithError(w, 401, "Expired token", nil)
		return
	}
	 if token.IsUsed {
		err = cfg.db.RevokeDelegationByID(ctx, token.DelegationID)
		if err != nil {
			respondWithError(w, 500, "Error revoking refresh session", err)
			return
		}
		err = cfg.db.ExpireAllDelegationTokens(ctx, token.DelegationID)
		if err != nil {
			respondWithError(w, 500, "Error expiring tokens of delegation", err)
		}

		respondWithError(w, 401, "Refresh token has already been used", nil)
		return
	}

	newJWT, err := auth.MakeJWT(token.UserID)
	if err != nil {
		respondWithError(w, 500, "Error creating JWT", err)
		return 
	}

	del, err := cfg.db.GetDelegation(ctx, token.DelegationID)
	if err != nil {
		respondWithError(w, 401, "Session delegation not found", err)
		return
	}

	newToken, err := auth.MakeRefreshToken(cfg.db, token.UserID, del)
	if err != nil {
		respondWithError(w, 500, "Error creating new refresh token", err)
		return
	}

	err = cfg.db.ExpireToken(ctx, tokenHash)
	if err != nil {
		respondWithError(w, 500, "Error expiring refresh token", err)
		return
	}

	response := models.RefreshResponse{
		RefreshToken: 	newToken,
		AccessToken: 	newJWT,
		TokenType: 		"Bearer",
	}

	respondWithJSON(w, 200, response)
}