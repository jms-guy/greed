package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/database"
)

type contextKey string

const (
    userIDKey contextKey = "userID"
    accountKey contextKey = "account"
)

//Middleware function to handle user authorization 
func (cfg *apiConfig) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, 400, "Bad token", err)
			return
		}

		id, err := auth.ValidateJWT(token, auth.JWTSecret)
		if err != nil {
			respondWithError(w, 401, "Invalid JWT", err)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//Middleware function to handle account authorization
func (cfg *apiConfig) AccountMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userIDValue := ctx.Value(userIDKey)
		userID, ok := userIDValue.(uuid.UUID)
		if !ok {
			respondWithError(w, 400, "Bad userID in context", nil)
			return
		}

		accID := r.PathValue("accountid")
		accountId, err := uuid.Parse(accID)
		if err != nil {
			respondWithError(w, 400, "Error parsing account ID", err)
			return
		}

		account, err := cfg.db.GetAccount(ctx, database.GetAccountParams{
			ID: accountId,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondWithError(w, 404, "Account not found", nil)
				return
			}
			respondWithError(w, 500, "Error getting account from database", err)
			return
		}

		ctx = context.WithValue(ctx, accountKey, account)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}