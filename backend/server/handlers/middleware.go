package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"
	"github.com/go-chi/chi/v5"
	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/encrypt"
)

type contextKey string

const (
    userIDKey 		contextKey = "userID"
	accessTokenKey 	contextKey = "accessTokenKey"
    accountKey 		contextKey = "account"
	requestIDKey 	contextKey = "requestID"
)

//Middleware function to handle user authorization.
//Serves following handlers with UserID in context
func (app *AppServer) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			app.respondWithError(w, 400, "Bad token", err)
			return
		}

		id, err := auth.ValidateJWT(app.Config, token)
		if err != nil {
			app.respondWithError(w, 401, "Invalid JWT", err)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//Middleware function to handle the Plaid access token.
//Serves following handlers with Plaid Access token in context
func (app *AppServer) AccessTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userIDValue := ctx.Value(userIDKey)
		userID, ok := userIDValue.(uuid.UUID)
		if !ok {
			app.respondWithError(w, 400, "Bad userID in context", nil)
			return
		}

		itemID := chi.URLParam(r, "item-id")
	
		token, err := app.Db.GetAccessToken(ctx, itemID)
		if err != nil {
			if err == sql.ErrNoRows {
				app.respondWithError(w, 400, "No item found for user", nil)
				return
			} 
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting access token: %w", err))
			return 
		}

		if token.UserID != userID {
			app.respondWithError(w, 403, "UserID does not match item's database record", nil)
			return
		}

		accessTokenbytes, err := encrypt.DecryptAccessToken(token.AccessToken, app.Config.AESKey)
		if err != nil {
			app.respondWithError(w, 500, "Error decrypting access token", err)
			return 
		}

		accessToken := string(accessTokenbytes)

		ctx = context.WithValue(ctx, accessTokenKey, accessToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


//Middleware function to handle account authorization.
//Serves following handlers with an account struct in context
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

		account, err := app.Db.GetAccountById(ctx, database.GetAccountByIdParams{
			ID: accID,
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

//Logging middleware function.
//Logs details of http request, and any errors
func LoggingMiddleware(Logger log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			requestID := uuid.New().String()

			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)

			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					Logger.Log(
						"requestID", requestID,
						"err", err,
						"trace", debug.Stack(),
					)
				}
			}()

			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			Logger.Log(
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

//Middleware for rate limiting based off IP address
func (app *AppServer) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.Logger.Log(
				"level", "error",
				"msg", "invalid IP",
				"err", err,
			)
			http.Error(w, "Invalid IP", 400)
			return 
		}

		limiter := app.Limiter.GetLimiter(ip, app.Config.RateLimit, app.Config.RateRefresh)
		if limiter.Allow() {
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			app.Logger.Log(
				"level", "trace",
				"msg", "rate limit exceeded",
				"ip", ip,
			)
			http.Error(w, "Rate Limit Exceeded", http.StatusTooManyRequests)
		}
	})
}

//Dev middleware for accessing admin endpoints
func (app *AppServer) DevAuthMiddleware(next http.Handler) http.Handler {
	isDev := app.Config.Environment == "dev"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isDev {
			next.ServeHTTP(w, r)
		} else {
			app.Logger.Log(
				"level", "warn",
				"msg", "attempted access to admin route in non-dev environment",
				"ip", r.RemoteAddr,
			)
			http.Error(w, "Not found", http.StatusForbidden)
		}
	})
}