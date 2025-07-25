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
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/server/handlers"
	"github.com/jms-guy/greed/models"
	"github.com/plaid/plaid-go/v36/plaid"
)

func TestHandlerGetLinkToken(t *testing.T) {
    tests := []struct {
        name            string
        userIDInContext uuid.UUID
		requestBody 	string
        mockDb          *mockDatabaseService
		mockPS 			*mockPlaidService
		mockAuth 		*mockAuthService
        expectedStatus  int 
        expectedBody    string  
    }{
		{
            name: "should successfully get a link token from plaid",
            userIDInContext: testUserID,
            mockDb: &mockDatabaseService{},
			mockPS: &mockPlaidService{
				GetLinkTokenFunc: func(ctx context.Context, userID, webhookURL string) (string, error) {
					return "link_token", nil
				},
			},
            expectedStatus: http.StatusOK,
            expectedBody: "link_token",
        },
		{
            name: "should err with bad userID in context",
            userIDInContext: uuid.Nil,
            mockDb: &mockDatabaseService{},
			mockPS: &mockPlaidService{
				GetLinkTokenFunc: func(ctx context.Context, userID, webhookURL string) (string, error) {
					return "link_token", nil
				},
			},
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad userID in context",
        },
		{
            name: "should err on getting link token from plaid",
            userIDInContext: testUserID,
            mockDb: &mockDatabaseService{},
			mockPS: &mockPlaidService{
				GetLinkTokenFunc: func(ctx context.Context, userID, webhookURL string) (string, error) {
					return "link_token", fmt.Errorf("mock error")
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Error getting link token from Plaid",
        },
	}

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/plaid/get-link-token", bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")

            ctx := context.WithValue(req.Context(), handlers.GetUserIDContextKey(), tt.userIDInContext)

            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
				PService: tt.mockPS,
                Logger: kitlog.NewNopLogger(),
				Config: &config.Config{PlaidWebhookURL: ""},
            }

            mockApp.HandlerGetLinkToken(rr, req)

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

func TestHandlerGetLinkTokenForUpdateMode(t *testing.T) {
    tests := []struct {
        name            string
        userIDInContext uuid.UUID
		accessTokenInContext any
		requestBody 	string
        mockDb          *mockDatabaseService
		mockPS 			*mockPlaidService
		mockAuth 		*mockAuthService
        expectedStatus  int 
        expectedBody    string  
    }{
		{
            name: "should successfully get a link token from plaid for update mode",
            userIDInContext: testUserID,
			accessTokenInContext: testAccessToken,
            mockDb: &mockDatabaseService{},
			mockPS: &mockPlaidService{
				GetLinkTokenForUpdateModeFunc: func(ctx context.Context, userID, accessToken, webhookURL string) (string, error) {
					return "link_token", nil
				},
			},
            expectedStatus: http.StatusOK,
            expectedBody: "link_token",
        },
		{
            name: "should err with bad userID",
            userIDInContext: uuid.Nil,
			accessTokenInContext: testAccessToken,
            mockDb: &mockDatabaseService{},
			mockPS: &mockPlaidService{
				GetLinkTokenForUpdateModeFunc: func(ctx context.Context, userID, accessToken, webhookURL string) (string, error) {
					return "link_token", nil
				},
			},
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad userID in context",
        },
		{
            name: "should err with bad access token",
            userIDInContext: testUserID,
			accessTokenInContext: 1,
            mockDb: &mockDatabaseService{},
			mockPS: &mockPlaidService{
				GetLinkTokenForUpdateModeFunc: func(ctx context.Context, userID, accessToken, webhookURL string) (string, error) {
					return "link_token", nil
				},
			},
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad access token in context",
        },
		{
            name: "should err on getting link token",
            userIDInContext: testUserID,
			accessTokenInContext: testAccessToken,
            mockDb: &mockDatabaseService{},
			mockPS: &mockPlaidService{
				GetLinkTokenForUpdateModeFunc: func(ctx context.Context, userID, accessToken, webhookURL string) (string, error) {
					return "link_token", fmt.Errorf("mock error")
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Service error",
        },
	}

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/plaid/get-link-token-update", bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")

            ctx := context.WithValue(req.Context(), handlers.GetUserIDContextKey(), tt.userIDInContext)
			ctx = context.WithValue(ctx, handlers.GetAccessTokenKey(), tt.accessTokenInContext)

            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
				PService: tt.mockPS,
                Logger: kitlog.NewNopLogger(),
				Config: &config.Config{PlaidWebhookURL: ""},
            }

            mockApp.HandlerGetLinkTokenForUpdateMode(rr, req)

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

func TestHandlerGetAccessToken(t *testing.T) {
    tests := []struct {
        name            string
        userIDInContext uuid.UUID
		requestBody 	string
        mockDb          *mockDatabaseService
		mockPS 			*mockPlaidService
		mockAuth 		*mockAuthService
		mockEncryptor 	*mockEncryptor
        expectedStatus  int 
        expectedBody    string  
    }{
		{
            name: "should successfully get a link token from plaid",
            userIDInContext: testUserID,
			requestBody: `{"public_token": "testToken", "nickname": "testName"}`,
            mockDb: &mockDatabaseService{
				CreateItemFunc: func(ctx context.Context, arg database.CreateItemParams) (database.PlaidItem, error) {
					return database.PlaidItem{}, nil
				},
			},
			mockPS: &mockPlaidService{
				GetAccessTokenFunc: func(ctx context.Context, publicToken string) (models.AccessResponse, error) {
					return models.AccessResponse{AccessToken: testAccessToken, ItemID: testItemID, InstitutionName: "testing", RequestID: "1"}, nil 
				},
			},
			mockEncryptor: &mockEncryptor{
				EncryptAccessTokenFunc: func(plaintext []byte, keyString string) (string, error) {
					return "encrypted", nil
				},
			},
            expectedStatus: http.StatusCreated,
            expectedBody: "Item created",
        },
		{
            name: "should err with bad userID in context",
            userIDInContext: uuid.Nil,
			requestBody: `{"public_token": "testToken", "nickname": "testName"}`,
            mockDb: &mockDatabaseService{
				CreateItemFunc: func(ctx context.Context, arg database.CreateItemParams) (database.PlaidItem, error) {
					return database.PlaidItem{}, nil
				},
			},
			mockPS: &mockPlaidService{
				GetAccessTokenFunc: func(ctx context.Context, publicToken string) (models.AccessResponse, error) {
					return models.AccessResponse{AccessToken: testAccessToken, ItemID: testItemID, InstitutionName: "testing", RequestID: "1"}, nil 
				},
			},
			mockEncryptor: &mockEncryptor{
				EncryptAccessTokenFunc: func(plaintext []byte, keyString string) (string, error) {
					return "encrypted", nil
				},
			},
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad userID in context",
        },
		{
            name: "should err on decoding request",
            userIDInContext: testUserID,
			requestBody: "",
            mockDb: &mockDatabaseService{
				CreateItemFunc: func(ctx context.Context, arg database.CreateItemParams) (database.PlaidItem, error) {
					return database.PlaidItem{}, nil
				},
			},
			mockPS: &mockPlaidService{
				GetAccessTokenFunc: func(ctx context.Context, publicToken string) (models.AccessResponse, error) {
					return models.AccessResponse{AccessToken: testAccessToken, ItemID: testItemID, InstitutionName: "testing", RequestID: "1"}, nil 
				},
			},
			mockEncryptor: &mockEncryptor{
				EncryptAccessTokenFunc: func(plaintext []byte, keyString string) (string, error) {
					return "encrypted", nil
				},
			},
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Couldn't decode JSON data",
        },
		{
            name: "should err on getting access token from Plaid",
            userIDInContext: testUserID,
			requestBody: `{"public_token": "testToken", "nickname": "testName"}`,
            mockDb: &mockDatabaseService{
				CreateItemFunc: func(ctx context.Context, arg database.CreateItemParams) (database.PlaidItem, error) {
					return database.PlaidItem{}, nil
				},
			},
			mockPS: &mockPlaidService{
				GetAccessTokenFunc: func(ctx context.Context, publicToken string) (models.AccessResponse, error) {
					return models.AccessResponse{AccessToken: testAccessToken, ItemID: testItemID, InstitutionName: "testing", RequestID: "1"}, fmt.Errorf("mock error") 
				},
			},
			mockEncryptor: &mockEncryptor{
				EncryptAccessTokenFunc: func(plaintext []byte, keyString string) (string, error) {
					return "encrypted", nil
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Error getting access token",
        },
		{
            name: "should err on encrypting access token",
            userIDInContext: testUserID,
			requestBody: `{"public_token": "testToken", "nickname": "testName"}`,
            mockDb: &mockDatabaseService{
				CreateItemFunc: func(ctx context.Context, arg database.CreateItemParams) (database.PlaidItem, error) {
					return database.PlaidItem{}, nil
				},
			},
			mockPS: &mockPlaidService{
				GetAccessTokenFunc: func(ctx context.Context, publicToken string) (models.AccessResponse, error) {
					return models.AccessResponse{AccessToken: testAccessToken, ItemID: testItemID, InstitutionName: "testing", RequestID: "1"}, nil 
				},
			},
			mockEncryptor: &mockEncryptor{
				EncryptAccessTokenFunc: func(plaintext []byte, keyString string) (string, error) {
					return "encrypted", fmt.Errorf("mock error")
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Error encrypting access token",
        },
		{
            name: "should err on creating item in database",
            userIDInContext: testUserID,
			requestBody: `{"public_token": "testToken", "nickname": "testName"}`,
            mockDb: &mockDatabaseService{
				CreateItemFunc: func(ctx context.Context, arg database.CreateItemParams) (database.PlaidItem, error) {
					return database.PlaidItem{}, fmt.Errorf("mock error")
				},
			},
			mockPS: &mockPlaidService{
				GetAccessTokenFunc: func(ctx context.Context, publicToken string) (models.AccessResponse, error) {
					return models.AccessResponse{AccessToken: testAccessToken, ItemID: testItemID, InstitutionName: "testing", RequestID: "1"}, nil 
				},
			},
			mockEncryptor: &mockEncryptor{
				EncryptAccessTokenFunc: func(plaintext []byte, keyString string) (string, error) {
					return "encrypted", nil
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
	}

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/plaid/get-access-token", bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")

            ctx := context.WithValue(req.Context(), handlers.GetUserIDContextKey(), tt.userIDInContext)

            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
				PService: tt.mockPS,
				Config: &config.Config{AESKey: "12345"},
                Logger: kitlog.NewNopLogger(),
				Encryptor: tt.mockEncryptor,
            }

            mockApp.HandlerGetAccessToken(rr, req)

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
		{
            name: "should err with bad access token in context",
            tokenInContext: 1,
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
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad access token in context",
        },
		{
            name: "should err on getting cursor",
            tokenInContext: testAccessToken,
            pathParams: map[string]string{"item-id": testItemID},
            mockDb: &mockDatabaseService{
                GetCursorFunc: func(ctx context.Context, arg database.GetCursorParams) (sql.NullString, error) {
					return sql.NullString{}, fmt.Errorf("mock error") 
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
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
		{
            name: "should err on getting transaction data from plaid",
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
					return []plaid.Transaction{}, []plaid.Transaction{}, []plaid.RemovedTransaction{}, "next_cursor", "requestID", fmt.Errorf("mock error") 
				},
			},
			mockTxnUpdater: &mockTxnUpdaterService{
				ApplyTransactionUpdatesFunc: func(ctx context.Context, added []plaid.Transaction, modified []plaid.Transaction, removed []plaid.RemovedTransaction, nextCursor string, itemID string) error {
					return nil
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Service error",
        },
		{
            name: "should err on applying transaction updates",
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
					return fmt.Errorf("mock error")
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
		{
            name: "should err on getting item by id",
            tokenInContext: testAccessToken,
            pathParams: map[string]string{"item-id": testItemID},
            mockDb: &mockDatabaseService{
                GetCursorFunc: func(ctx context.Context, arg database.GetCursorParams) (sql.NullString, error) {
					return sql.NullString{String: "cursor", Valid: true}, nil 
				}, 
				GetItemByIDFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{}, fmt.Errorf("mock error")
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
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
		{
            name: "should err on getting transactions for user",
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
					return []database.Transaction{}, fmt.Errorf("mock error")
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
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
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

func TestHandlerUpdateBalances(t *testing.T) {
    tests := []struct {
        name            string
        tokenInContext  any
        pathParams      map[string]string
		requestBody 	string
        mockDb          *mockDatabaseService
		mockPS 			*mockPlaidService
		mockAuth 		*mockAuthService
        expectedStatus  int 
        expectedBody    string  
    }{
		{
            name: "should successfully update account balances",
            tokenInContext: testAccessToken,
            pathParams: map[string]string{"item-id": testItemID},
            mockDb: &mockDatabaseService{
                UpdateBalancesFunc: func(ctx context.Context, arg database.UpdateBalancesParams) (database.Account, error) {
					return database.Account{ID: testAccountID}, nil
				},
            },
			mockPS: &mockPlaidService{
				GetBalancesFunc: func(ctx context.Context, accessToken string) (plaid.AccountsGetResponse, string, error) {
					return plaid.AccountsGetResponse{Accounts:[]plaid.AccountBase{{AccountId: testAccountID}}}, "requestID", nil
				},
			},
            expectedStatus: http.StatusOK,
            expectedBody: testAccountID,
        },
		{
            name: "should err with bad access token in context",
            tokenInContext: 1,
            pathParams: map[string]string{"item-id": testItemID},
            mockDb: &mockDatabaseService{
                UpdateBalancesFunc: func(ctx context.Context, arg database.UpdateBalancesParams) (database.Account, error) {
					return database.Account{ID: testAccountID}, nil
				},
            },
			mockPS: &mockPlaidService{
				GetBalancesFunc: func(ctx context.Context, accessToken string) (plaid.AccountsGetResponse, string, error) {
					return plaid.AccountsGetResponse{Accounts:[]plaid.AccountBase{{AccountId: testAccountID}}}, "requestID", nil
				},
			},
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad access token in context",
        },
		{
            name: "should err with plaid service error on getting balances",
            tokenInContext: testAccessToken,
            pathParams: map[string]string{"item-id": testItemID},
            mockDb: &mockDatabaseService{
                UpdateBalancesFunc: func(ctx context.Context, arg database.UpdateBalancesParams) (database.Account, error) {
					return database.Account{ID: testAccountID}, nil
				},
            },
			mockPS: &mockPlaidService{
				GetBalancesFunc: func(ctx context.Context, accessToken string) (plaid.AccountsGetResponse, string, error) {
					return plaid.AccountsGetResponse{Accounts:[]plaid.AccountBase{{AccountId: testAccountID}}}, "requestID", fmt.Errorf("mock error")
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Service error",
        },
		{
            name: "should err on updating balances in database",
            tokenInContext: testAccessToken,
            pathParams: map[string]string{"item-id": testItemID},
            mockDb: &mockDatabaseService{
                UpdateBalancesFunc: func(ctx context.Context, arg database.UpdateBalancesParams) (database.Account, error) {
					return database.Account{ID: testAccountID}, fmt.Errorf("mock error")
				},
            },
			mockPS: &mockPlaidService{
				GetBalancesFunc: func(ctx context.Context, accessToken string) (plaid.AccountsGetResponse, string, error) {
					return plaid.AccountsGetResponse{Accounts:[]plaid.AccountBase{{AccountId: testAccountID}}}, "requestID", nil
				},
			},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
	}

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("PUT", fmt.Sprintf("/api/items/%s/access/balances", tt.pathParams["item-id"]), bytes.NewBufferString(tt.requestBody))
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
            }

            mockApp.HandlerUpdateBalances(rr, req)

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