package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"runtime/debug"
	"time"
	"github.com/go-chi/chi/v5"
	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/database"
)

type contextKey string

const (
    userIDKey contextKey = "userID"
    accountKey contextKey = "account"
	requestIDKey contextKey = "requestID"
)

//Middleware function to handle user authorization 
func (app *AppServer) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			app.respondWithError(w, 400, "Bad token", err)
			return
		}

		id, err := auth.ValidateJWT(app.config, token)
		if err != nil {
			app.respondWithError(w, 401, "Invalid JWT", err)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//Middleware function to handle account authorization
func (app *AppServer) AccountMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userIDValue := ctx.Value(userIDKey)
		userID, ok := userIDValue.(uuid.UUID)
		if !ok {
			app.respondWithError(w, 400, "Bad userID in context", nil)
			return
		}

		accID := chi.URLParam(r, "accountid")
		accountId, err := uuid.Parse(accID)
		if err != nil {
			app.respondWithError(w, 400, "Error parsing account ID", err)
			return
		}

		account, err := app.db.GetAccount(ctx, database.GetAccountParams{
			ID: accountId,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				app.respondWithError(w, 404, "Account not found", nil)
				return
			}
			app.respondWithError(w, 500, "Error getting account from database", err)
			return
		}

		ctx = context.WithValue(ctx, accountKey, account)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//Below is logging middleware - includes panic recovery 
type responseWriter struct{
	http.ResponseWriter
	status			int
	wroteHeader		bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (rw *responseWriter) Write(b []byte) (int, error) {
    if !rw.wroteHeader {
        rw.WriteHeader(http.StatusOK)
    }
    return rw.ResponseWriter.Write(b)
}

func LoggingMiddleware(logger log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			requestID := uuid.New().String()

			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)

			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Log(
						"requestID", requestID,
						"err", err,
						"trace", debug.Stack(),
					)
				}
			}()

			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			logger.Log(
				"requestID", requestID,
				"status", wrapped.status,
				"method", r.Method,
				"path", r.URL.EscapedPath(),
				"duration", time.Since(start),
			)
		}

		return http.HandlerFunc(fn)
	}
}