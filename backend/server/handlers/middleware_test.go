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

func TestAuthMiddleware(t *testing.T) {
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

            mockApp.AuthMiddleware(rr, req)

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