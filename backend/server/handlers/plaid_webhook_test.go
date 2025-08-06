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

	kitlog "github.com/go-kit/log"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/server/handlers"
	"github.com/plaid/plaid-go/v36/plaid"
)

func TestHandlerPlaidWebhook(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      string
		mockDb           *mockDatabaseService
		mockAuth         *mockAuthService
		mockPlaidService *mockPlaidService
		expectedStatus   int
		expectedBody     string
	}{
		{
			name:        "should successfully get webhook and create record",
			requestBody: `{"webhook_type":"testType", "webhook_code":"testCode", "item_id":"12345"}`,
			mockDb: &mockDatabaseService{
				GetItemByIDFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{ID: "12345"}, nil
				},
				CreatePlaidWebhookRecordFunc: func(ctx context.Context, arg database.CreatePlaidWebhookRecordParams) (database.PlaidWebhookRecord, error) {
					return database.PlaidWebhookRecord{}, nil
				},
			},
			mockAuth: &mockAuthService{
				VerifyPlaidJWTFunc: func(p auth.PlaidKeyFetcher, ctx context.Context, tokenString string) error {
					return nil
				},
			},
			mockPlaidService: &mockPlaidService{
				GetWebhookVerificationKeyFunc: func(ctx context.Context, keyID string) (plaid.JWKPublicKey, error) {
					return plaid.JWKPublicKey{}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "should err with bad request",
			mockDb: &mockDatabaseService{
				GetItemByIDFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{ID: "12345"}, nil
				},
				CreatePlaidWebhookRecordFunc: func(ctx context.Context, arg database.CreatePlaidWebhookRecordParams) (database.PlaidWebhookRecord, error) {
					return database.PlaidWebhookRecord{}, nil
				},
			},
			mockAuth: &mockAuthService{
				VerifyPlaidJWTFunc: func(p auth.PlaidKeyFetcher, ctx context.Context, tokenString string) error {
					return nil
				},
			},
			mockPlaidService: &mockPlaidService{
				GetWebhookVerificationKeyFunc: func(ctx context.Context, keyID string) (plaid.JWKPublicKey, error) {
					return plaid.JWKPublicKey{}, nil
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad request",
		},
		{
			name:        "should err with Unauthorized",
			requestBody: `{"webhook_type":"testType", "webhook_code":"testCode", "item_id":"12345"}`,
			mockDb: &mockDatabaseService{
				GetItemByIDFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{ID: "12345"}, nil
				},
				CreatePlaidWebhookRecordFunc: func(ctx context.Context, arg database.CreatePlaidWebhookRecordParams) (database.PlaidWebhookRecord, error) {
					return database.PlaidWebhookRecord{}, nil
				},
			},
			mockAuth: &mockAuthService{
				VerifyPlaidJWTFunc: func(p auth.PlaidKeyFetcher, ctx context.Context, tokenString string) error {
					return fmt.Errorf("mock error")
				},
			},
			mockPlaidService: &mockPlaidService{
				GetWebhookVerificationKeyFunc: func(ctx context.Context, keyID string) (plaid.JWKPublicKey, error) {
					return plaid.JWKPublicKey{}, nil
				},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:        "should err with item not found",
			requestBody: `{"webhook_type":"testType", "webhook_code":"testCode", "item_id":"12345"}`,
			mockDb: &mockDatabaseService{
				GetItemByIDFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{}, sql.ErrNoRows
				},
				CreatePlaidWebhookRecordFunc: func(ctx context.Context, arg database.CreatePlaidWebhookRecordParams) (database.PlaidWebhookRecord, error) {
					return database.PlaidWebhookRecord{}, nil
				},
			},
			mockAuth: &mockAuthService{
				VerifyPlaidJWTFunc: func(p auth.PlaidKeyFetcher, ctx context.Context, tokenString string) error {
					return nil
				},
			},
			mockPlaidService: &mockPlaidService{
				GetWebhookVerificationKeyFunc: func(ctx context.Context, keyID string) (plaid.JWKPublicKey, error) {
					return plaid.JWKPublicKey{}, nil
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not Found",
		},
		{
			name:        "should err with Internal server error, on getting item from database",
			requestBody: `{"webhook_type":"testType", "webhook_code":"testCode", "item_id":"12345"}`,
			mockDb: &mockDatabaseService{
				GetItemByIDFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{}, fmt.Errorf("mock error")
				},
				CreatePlaidWebhookRecordFunc: func(ctx context.Context, arg database.CreatePlaidWebhookRecordParams) (database.PlaidWebhookRecord, error) {
					return database.PlaidWebhookRecord{}, nil
				},
			},
			mockAuth: &mockAuthService{
				VerifyPlaidJWTFunc: func(p auth.PlaidKeyFetcher, ctx context.Context, tokenString string) error {
					return nil
				},
			},
			mockPlaidService: &mockPlaidService{
				GetWebhookVerificationKeyFunc: func(ctx context.Context, keyID string) (plaid.JWKPublicKey, error) {
					return plaid.JWKPublicKey{}, nil
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error",
		},
		{
			name:        "should err creating webhook record",
			requestBody: `{"webhook_type":"testType", "webhook_code":"testCode", "item_id":"12345"}`,
			mockDb: &mockDatabaseService{
				GetItemByIDFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{ID: "12345"}, nil
				},
				CreatePlaidWebhookRecordFunc: func(ctx context.Context, arg database.CreatePlaidWebhookRecordParams) (database.PlaidWebhookRecord, error) {
					return database.PlaidWebhookRecord{}, fmt.Errorf("mock error")
				},
			},
			mockAuth: &mockAuthService{
				VerifyPlaidJWTFunc: func(p auth.PlaidKeyFetcher, ctx context.Context, tokenString string) error {
					return nil
				},
			},
			mockPlaidService: &mockPlaidService{
				GetWebhookVerificationKeyFunc: func(ctx context.Context, keyID string) (plaid.JWKPublicKey, error) {
					return plaid.JWKPublicKey{}, nil
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/plaid-webhook", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("plaid-verification", "testToken")

			rr := httptest.NewRecorder()

			mockApp := &handlers.AppServer{
				Db:       tt.mockDb,
				Auth:     tt.mockAuth,
				PService: tt.mockPlaidService,
				Logger:   kitlog.NewNopLogger(),
			}

			mockApp.HandlerPlaidWebhook(rr, req)

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
