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
	"github.com/plaid/plaid-go/v36/plaid"
)


func TestHandlerCreateAccounts(t *testing.T) {
    tests := []struct {
        name            string
		userIDInContext uuid.UUID
        tokenInContext  any
        pathParams      map[string]string
		requestBody 	string
        mockDb          *mockDatabaseService
		mockPS 			*mockPlaidService
        expectedStatus  int 
        expectedBody    string  
    }{
		{
            name: "should successfully create account records for item",
            userIDInContext: testUserID,
            tokenInContext: testAccessToken,
            pathParams: map[string]string{"item-id": "12345"},
            mockDb: &mockDatabaseService{
                CreateAccountFunc: func(ctx context.Context, arg database.CreateAccountParams) (database.Account, error) {
					return testAccount, nil
				}, 
            },
			mockPS: &mockPlaidService{
				GetAccountsFunc: func(ctx context.Context, accessToken string) ([]plaid.AccountBase, string, error) {
					return []plaid.AccountBase{{AccountId: testAccountID}}, "", nil 
				},
			},
            expectedStatus: http.StatusCreated,
            expectedBody: testAccountID,
        },
		{
            name: "should err with bad userId in context",
            userIDInContext: uuid.Nil,
            tokenInContext: testAccessToken,
            pathParams: map[string]string{"item-id": "12345"},
            mockDb: &mockDatabaseService{
                CreateAccountFunc: func(ctx context.Context, arg database.CreateAccountParams) (database.Account, error) {
					return testAccount, nil
				}, 
            },
			mockPS: &mockPlaidService{
				GetAccountsFunc: func(ctx context.Context, accessToken string) ([]plaid.AccountBase, string, error) {
					return []plaid.AccountBase{{AccountId: testAccountID}}, "", nil 
				},
			},
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad userID in context",
        },
		{
            name: "should err with bad access token in context",
            userIDInContext: testUserID,
            tokenInContext: 1,
            pathParams: map[string]string{"item-id": "12345"},
            mockDb: &mockDatabaseService{
                CreateAccountFunc: func(ctx context.Context, arg database.CreateAccountParams) (database.Account, error) {
					return testAccount, nil
				}, 
            },
			mockPS: &mockPlaidService{
				GetAccountsFunc: func(ctx context.Context, accessToken string) ([]plaid.AccountBase, string, error) {
					return []plaid.AccountBase{{AccountId: testAccountID}}, "", nil 
				},
			},
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad access token in context",
        },
		{
            name: "should err on getting data from Plaid api",
            userIDInContext: testUserID,
            tokenInContext: testAccessToken,
            pathParams: map[string]string{"item-id": "12345"},
            mockDb: &mockDatabaseService{
                CreateAccountFunc: func(ctx context.Context, arg database.CreateAccountParams) (database.Account, error) {
					return testAccount, nil
				}, 
            },
			mockPS: &mockPlaidService{
				GetAccountsFunc: func(ctx context.Context, accessToken string) ([]plaid.AccountBase, string, error) {
					return []plaid.AccountBase{{AccountId: testAccountID}}, "", fmt.Errorf("mock error") 
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Service Error",
        },
		{
            name: "should err on creating account in database",
            userIDInContext: testUserID,
            tokenInContext: testAccessToken,
            pathParams: map[string]string{"item-id": "12345"},
            mockDb: &mockDatabaseService{
                CreateAccountFunc: func(ctx context.Context, arg database.CreateAccountParams) (database.Account, error) {
					return database.Account{}, fmt.Errorf("mock error")
				}, 
            },
			mockPS: &mockPlaidService{
				GetAccountsFunc: func(ctx context.Context, accessToken string) ([]plaid.AccountBase, string, error) {
					return []plaid.AccountBase{{AccountId: testAccountID}}, "", nil 
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
    }

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", fmt.Sprintf("/api/items/%s/access/accounts", tt.pathParams["item-id"]), bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            
            rctx := chi.NewRouteContext()
            for key, value := range tt.pathParams {
                rctx.URLParams.Add(key, value)
            }

            ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
            ctx = context.WithValue(ctx,handlers.GetUserIDContextKey(), tt.userIDInContext)
            ctx = context.WithValue(ctx, handlers.GetAccessTokenKey(), tt.tokenInContext)

            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
				PService: tt.mockPS,
                Logger: kitlog.NewNopLogger(),
            }

            mockApp.HandlerCreateAccounts(rr, req)

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

func TestHandlerSyncTransactions(t *testing.T) {
    tests := []struct {
        name            string
        tokenInContext  any
        pathParams      map[string]string
		requestBody 	string
        mockDb          *mockDatabaseService
		mockPS 			*mockPlaidService
		mockAuth 		*mockAuthService
		mockTxnUpdater  *mockTxnUpdaterService
        expectedStatus  int 
        expectedBody    string  
    }{
		{
            name: "should successfully sync database transaction data with fetched Plaid data",
            tokenInContext: testAccessToken,
            pathParams: map[string]string{"item-id": testItemID},
            mockDb: &mockDatabaseService{
                GetCursorFunc: func(ctx context.Context, arg database.GetCursorParams) (sql.NullString, error) {
					return sql.NullString{String: "cursor", Valid: true}, nil 
				}, 
				GetItemByIDFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{ID: testItemID}, nil
				},
				GetTransactionsForUserFunc: func(ctx context.Context, userID uuid.UUID) ([]database.Transaction, error) {
					return []database.Transaction{{AccountID: testAccountID}}, nil
				},
            },
			mockPS: &mockPlaidService{
				GetTransactionsFunc: func(ctx context.Context, accessToken string, cursor string) (added []plaid.Transaction, modified []plaid.Transaction, removed []plaid.RemovedTransaction, nextCursor string, reqID string, err error) {
					return []plaid.Transaction{{AccountId: testAccountID}}, []plaid.Transaction{}, []plaid.RemovedTransaction{}, "next_cursor", "requestID", nil 
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ApplyTransactionUpdatesFunc: func(ctx context.Context, added []plaid.Transaction, modified []plaid.Transaction, removed []plaid.RemovedTransaction, nextCursor string, itemID string) error {
					return nil
				},
			},
            expectedStatus: http.StatusOK,
            expectedBody: testAccountID,
        },
	}

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", fmt.Sprintf("/api/items/%s/access/transactions", tt.pathParams["item-id"]), bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            
            rctx := chi.NewRouteContext()
            for key, value := range tt.pathParams {
                rctx.URLParams.Add(key, value)
            }

            ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
            ctx = context.WithValue(ctx, handlers.GetAccessTokenKey(), tt.tokenInContext)

            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
				PService: tt.mockPS,
                Logger: kitlog.NewNopLogger(),
				TxnUpdater: tt.mockTxnUpdater,
            }

            mockApp.HandlerSyncTransactions(rr, req)

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