package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/server/handlers"
)


func TestHandlerRefreshToken(t *testing.T) {
    tests := []struct {
        name            string
        userIDInContext uuid.UUID
		requestBody 	string
        mockDb          *mockDatabaseService
		mockAuth 		*mockAuthService
		mockTxnUpdater  *mockTxnUpdaterService
        expectedStatus  int 
        expectedBody    string  
    }{
		{
            name: "should successfully create a refresh token",
            userIDInContext: testUserID,
			requestBody: `{"refresh_token": "token"}`,
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(time.Hour)}, nil
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return nil
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", nil
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", nil
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return nil
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return nil
				},
			},
            expectedStatus: http.StatusOK,
            expectedBody: "JWT",
        },
		{
            name: "should err on decoding request",
            userIDInContext: testUserID,
			requestBody: "",
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(time.Hour)}, nil
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return nil
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", nil
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", nil
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return nil
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return nil
				},
			},
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Error decoding JSON parameters",
        },
		{
            name: "should err with no refresh token found",
            userIDInContext: testUserID,
			requestBody: `{"refresh_token": "token"}`,
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(time.Hour)}, sql.ErrNoRows
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return nil
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", nil
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", nil
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return nil
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return nil
				},
			},
            expectedStatus: http.StatusUnauthorized,
            expectedBody: "Refresh token not found",
        },
		{
            name: "should err on getting refresh token",
            userIDInContext: testUserID,
			requestBody: `{"refresh_token": "token"}`,
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(time.Hour)}, fmt.Errorf("mock error")
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return nil
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", nil
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", nil
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return nil
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return nil
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
		{
            name: "should err expiring expired delegation token",
            userIDInContext: testUserID,
			requestBody: `{"refresh_token": "token"}`,
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(-1*time.Hour)}, nil
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return nil
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", nil
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", nil
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return fmt.Errorf("mock error")
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return nil
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
		{
            name: "should err with expired delegation token",
            userIDInContext: testUserID,
			requestBody: `{"refresh_token": "token"}`,
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(-1*time.Hour)}, nil
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return nil
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", nil
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", nil
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return nil
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return nil
				},
			},
            expectedStatus: http.StatusUnauthorized,
            expectedBody: "Token is expired",
        },
		{
            name: "should err on revoking delegation",
            userIDInContext: testUserID,
			requestBody: `{"refresh_token": "token"}`,
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(time.Hour), IsUsed: true}, nil
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return nil
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", nil
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", nil
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return nil
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return fmt.Errorf("mock error")
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
		{
            name: "should err with revoked delegation",
            userIDInContext: testUserID,
			requestBody: `{"refresh_token": "token"}`,
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(time.Hour), IsUsed: true}, nil
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return nil
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", nil
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", nil
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return nil
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return nil
				},
			},
            expectedStatus: http.StatusUnauthorized,
            expectedBody: "Refresh token has already been used",
        },
		{
            name: "should err on creating JWT",
            userIDInContext: testUserID,
			requestBody: `{"refresh_token": "token"}`,
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(time.Hour)}, nil
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return nil
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", fmt.Errorf("mock error")
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", nil
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return nil
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return nil
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Error creating JWT",
        },
		{
            name: "should err with delegation not found",
            userIDInContext: testUserID,
			requestBody: `{"refresh_token": "token"}`,
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(time.Hour)}, nil
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return nil
				},
				GetDelegationFunc: func(ctx context.Context, id uuid.UUID) (database.Delegation, error) {
					return database.Delegation{}, sql.ErrNoRows
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", nil
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", nil
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return nil
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return nil
				},
			},
            expectedStatus: http.StatusUnauthorized,
            expectedBody: "Session delegation not found",
        },
		{
            name: "should err on getting delegation",
            userIDInContext: testUserID,
			requestBody: `{"refresh_token": "token"}`,
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(time.Hour)}, nil
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return nil
				},
				GetDelegationFunc: func(ctx context.Context, id uuid.UUID) (database.Delegation, error) {
					return database.Delegation{}, fmt.Errorf("mock error")
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", nil
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", nil
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return nil
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return nil
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
		{
            name: "should err on making refresh token",
            userIDInContext: testUserID,
			requestBody: `{"refresh_token": "token"}`,
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(time.Hour)}, nil
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return nil
				},
				GetDelegationFunc: func(ctx context.Context, id uuid.UUID) (database.Delegation, error) {
					return database.Delegation{}, nil
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", nil
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", fmt.Errorf("mock error")
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return nil
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return nil
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Error creating new refresh token",
        },
		{
            name: "should err on expiring refresh token",
            userIDInContext: testUserID,
			requestBody: `{"refresh_token": "token"}`,
            mockDb: &mockDatabaseService{
				GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
					return database.RefreshToken{ExpiresAt: time.Now().Add(time.Hour)}, nil
				},
				ExpireTokenFunc: func(ctx context.Context, hashedToken string) error {
					return fmt.Errorf("mock error")
				},
				GetDelegationFunc: func(ctx context.Context, id uuid.UUID) (database.Delegation, error) {
					return database.Delegation{}, nil
				},
			},
			mockAuth: &mockAuthService{
				HashRefreshTokenFunc: func(token string) string {
					return "hashed"
				},
				MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
					return "JWT", nil
				},
				MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
					return "", nil
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ExpireDelegationFunc: func(ctx context.Context, tokenHash string, token database.RefreshToken) error {
					return nil
				},
				RevokeDelegationFunc: func(ctx context.Context, token database.RefreshToken) error {
					return nil
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
	}

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
				Auth: tt.mockAuth,
				TxnUpdater: tt.mockTxnUpdater,
				Config: &config.Config{},
                Logger: kitlog.NewNopLogger(),
            }

            mockApp.HandlerRefreshToken(rr, req)

            // --- Assertions ---
            if status := rr.Code; status != tt.expectedStatus {
                t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, tt.expectedStatus, rr.Body.String())
            }
            if !strings.Contains(rr.Body.String(), tt.expectedBody) {
                t.Errorf("handler returned unexpected body: got %s want body to contain %s", rr.Body.String(), tt.expectedBody)
            }
        })
    }
}
