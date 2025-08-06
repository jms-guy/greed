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

	"github.com/go-chi/chi/v5"
	kitlog "github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/server/handlers"
)

func TestHandlerGetAccountsForUser(t *testing.T) {
	tests := []struct {
		name            string
		userIDInContext uuid.UUID
		pathParams      map[string]string
		requestBody     string
		mockDb          *mockDatabaseService
		mockAuth        *mockAuthService
		expectedStatus  int
		expectedBody    string
	}{
		{
			name:            "should successfully get account records for user",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetAllAccountsForUserFunc: func(ctx context.Context, userID uuid.UUID) ([]database.Account, error) {
					return []database.Account{{ID: testAccountID, Name: testAccountName}}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   testAccountID,
		},
		{
			name:            "should err with bad userID in context",
			userIDInContext: uuid.Nil,
			mockDb: &mockDatabaseService{
				GetAllAccountsForUserFunc: func(ctx context.Context, userID uuid.UUID) ([]database.Account, error) {
					t.Fatalf("should not be called on err")
					return []database.Account{{ID: testAccountID, Name: testAccountName}}, nil
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad userID in context",
		},
		{
			name:            "should err with no accounts found",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetAllAccountsForUserFunc: func(ctx context.Context, userID uuid.UUID) ([]database.Account, error) {
					return []database.Account{}, sql.ErrNoRows
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "No accounts found for user",
		},
		{
			name:            "should err on getting accounts for user",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetAllAccountsForUserFunc: func(ctx context.Context, userID uuid.UUID) ([]database.Account, error) {
					return []database.Account{}, fmt.Errorf("mock error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/accounts", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			for key, value := range tt.pathParams {
				rctx.URLParams.Add(key, value)
			}

			ctx := context.WithValue(req.Context(), handlers.GetUserIDContextKey(), tt.userIDInContext)

			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			mockApp := &handlers.AppServer{
				Db:     tt.mockDb,
				Logger: kitlog.NewNopLogger(),
			}

			mockApp.HandlerGetAccountsForUser(rr, req)

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

func TestHandlerGetAccountData(t *testing.T) {
	tests := []struct {
		name             string
		accountInContext any
		pathParams       map[string]string
		requestBody      string
		mockDb           *mockDatabaseService
		mockAuth         *mockAuthService
		expectedStatus   int
		expectedBody     string
	}{
		{
			name:             "should successfully get record for single account",
			accountInContext: testAccount,
			pathParams:       map[string]string{"account": testAccountID},
			mockDb:           &mockDatabaseService{},
			expectedStatus:   http.StatusOK,
			expectedBody:     testAccountID,
		},
		{
			name:             "should err with bad account in context",
			accountInContext: "",
			pathParams:       map[string]string{"account": testAccountID},
			mockDb:           &mockDatabaseService{},
			expectedStatus:   http.StatusBadRequest,
			expectedBody:     "Bad account in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/accounts/%s/data", tt.pathParams["account-id"]), bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			for key, value := range tt.pathParams {
				rctx.URLParams.Add(key, value)
			}

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, handlers.GetAccountKey(), tt.accountInContext)

			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			mockApp := &handlers.AppServer{
				Db:     tt.mockDb,
				Logger: kitlog.NewNopLogger(),
			}

			mockApp.HandlerGetAccountData(rr, req)
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

func TestHandlerDeleteAccount(t *testing.T) {
	tests := []struct {
		name             string
		userIdInContext  uuid.UUID
		accountInContext any
		pathParams       map[string]string
		requestBody      string
		mockDb           *mockDatabaseService
		mockAuth         *mockAuthService
		expectedStatus   int
		expectedBody     string
	}{
		{
			name:             "should successfully delete an account record",
			userIdInContext:  testUserID,
			accountInContext: testAccount,
			pathParams:       map[string]string{"account": testAccountID},
			mockDb: &mockDatabaseService{
				DeleteAccountFunc: func(ctx context.Context, arg database.DeleteAccountParams) error {
					return nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Account deleted successfully",
		},
		{
			name:             "should err with bad userID in context",
			userIdInContext:  uuid.Nil,
			accountInContext: testAccount,
			pathParams:       map[string]string{"account": testAccountID},
			mockDb: &mockDatabaseService{
				DeleteAccountFunc: func(ctx context.Context, arg database.DeleteAccountParams) error {
					return nil
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad userID in context",
		},
		{
			name:             "should err with bad account in context",
			userIdInContext:  testUserID,
			accountInContext: "",
			pathParams:       map[string]string{"account": testAccountID},
			mockDb: &mockDatabaseService{
				DeleteAccountFunc: func(ctx context.Context, arg database.DeleteAccountParams) error {
					return nil
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad account in context",
		},
		{
			name:             "should err on deleting account record",
			userIdInContext:  testUserID,
			accountInContext: testAccount,
			pathParams:       map[string]string{"account": testAccountID},
			mockDb: &mockDatabaseService{
				DeleteAccountFunc: func(ctx context.Context, arg database.DeleteAccountParams) error {
					return fmt.Errorf("mock error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/accounts/%s", tt.pathParams["account-id"]), bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			for key, value := range tt.pathParams {
				rctx.URLParams.Add(key, value)
			}

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, handlers.GetUserIDContextKey(), tt.userIdInContext)
			ctx = context.WithValue(ctx, handlers.GetAccountKey(), tt.accountInContext)

			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			mockApp := &handlers.AppServer{
				Db:     tt.mockDb,
				Logger: kitlog.NewNopLogger(),
			}

			mockApp.HandlerDeleteAccount(rr, req)
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
