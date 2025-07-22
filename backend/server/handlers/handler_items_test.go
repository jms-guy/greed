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

func TestHandlerGetItems(t *testing.T) {
    tests := []struct {
        name            string
		userIDInContext uuid.UUID
        requestBody     string
        mockDb          *mockDatabaseService
		mockAuth 		*mockAuthService
        expectedStatus  int 
        expectedBody    string  
    }{
		{
            name: "should successfully return items for user",
			userIDInContext: testUserID,
            mockDb: &mockDatabaseService{
                GetItemsByUserFunc: func(ctx context.Context, userID uuid.UUID) ([]database.PlaidItem, error) {
					return []database.PlaidItem{{InstitutionName: testItemName}}, nil
				},
            },
            expectedStatus: http.StatusOK,
            expectedBody: testItemName,
        },
        {
            name: "should err with bad userID in context",
			userIDInContext: uuid.Nil,
            mockDb: &mockDatabaseService{},
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad userID in context",
        },
        {
            name: "should err with no items found",
			userIDInContext: testUserID,
            mockDb: &mockDatabaseService{
                GetItemsByUserFunc: func(ctx context.Context, userID uuid.UUID) ([]database.PlaidItem, error) {
					return []database.PlaidItem{}, sql.ErrNoRows
				},
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: "No items found for user",
        },
        {
            name: "should err on getting items for user",
			userIDInContext: testUserID,
            mockDb: &mockDatabaseService{
                GetItemsByUserFunc: func(ctx context.Context, userID uuid.UUID) ([]database.PlaidItem, error) {
					return []database.PlaidItem{}, fmt.Errorf("mock error")
				},
            },
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
	}

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/api/items", bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(),handlers.GetUserIDContextKey(), tt.userIDInContext)
            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
                Logger: kitlog.NewNopLogger(),
            }

            mockApp.HandlerGetItems(rr, req)

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

func TestHandlerUpdateItemName(t *testing.T) {
    tests := []struct {
        name            string
		userIDInContext uuid.UUID
        pathParams      map[string]string
        requestBody     string
        mockDb          *mockDatabaseService
		mockAuth 		*mockAuthService
        expectedStatus  int 
        expectedBody    string  
    }{
		{
            name: "should successfully update an item's nickname field",
			userIDInContext: testUserID,
            pathParams: map[string]string{"item-id": "1234"},
            requestBody: `{"nickname": "new_name"}`,
            mockDb: &mockDatabaseService{
                UpdateNicknameFunc: func(ctx context.Context, arg database.UpdateNicknameParams) error {
                    return nil
                },  
            },
            expectedStatus: http.StatusOK,
            expectedBody: "Item name updated successfully",
        },
        {
            name: "should err with bad userID in context",
			userIDInContext: uuid.Nil,
            pathParams: map[string]string{"item-id": "1234"},
            requestBody: `{"nickname": "new_name"}`,
            mockDb: &mockDatabaseService{
                UpdateNicknameFunc: func(ctx context.Context, arg database.UpdateNicknameParams) error {
                    return nil
                },  
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad userID in context",
        },
        {
            name: "should err on decoding request",
			userIDInContext: testUserID,
            pathParams: map[string]string{"item-id": "1234"},
            requestBody: "",
            mockDb: &mockDatabaseService{
                UpdateNicknameFunc: func(ctx context.Context, arg database.UpdateNicknameParams) error {
                    return nil
                },  
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad request data",
        },
        {
            name: "should err on updating item nickname",
			userIDInContext: testUserID,
            pathParams: map[string]string{"item-id": "1234"},
            requestBody: `{"nickname": "new_name"}`,
            mockDb: &mockDatabaseService{
                UpdateNicknameFunc: func(ctx context.Context, arg database.UpdateNicknameParams) error {
                    return fmt.Errorf("mock error")
                },  
            },
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
    }

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("PUT", fmt.Sprintf("/api/items/%s/name", tt.pathParams["item-id"]), bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            
            rctx := chi.NewRouteContext()
            for key, value := range tt.pathParams {
                rctx.URLParams.Add(key, value)
            }

            ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
            ctx = context.WithValue(ctx,handlers.GetUserIDContextKey(), tt.userIDInContext)

            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
                Logger: kitlog.NewNopLogger(),
            }

            mockApp.HandlerUpdateItemName(rr, req)

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

func TestHandlerDeleteItem(t *testing.T) {
    tests := []struct {
        name            string
		userIDInContext uuid.UUID
        accessTokenInCtx string
        pathParams      map[string]string
        requestBody     string
        mockDb          *mockDatabaseService
		mockAuth 		*mockAuthService
        mockPlaidService *mockPlaidService
        expectedStatus  int 
        expectedBody    string  
    }{
		{
            name: "should successfully delete an item record",
			userIDInContext: testUserID,
            accessTokenInCtx: testAccessToken,
            pathParams: map[string]string{"item-id": "12345"},
            mockDb: &mockDatabaseService{
                DeleteItemFunc: func(ctx context.Context, arg database.DeleteItemParams) error {
                    return nil 
                },  
            },
            mockPlaidService: &mockPlaidService{
                RemoveItemFunc: func(ctx context.Context, accessToken string) error {
                    return nil
                },
            },
            expectedStatus: http.StatusOK,
            expectedBody: "Item deleted successfully",
        },
        {
            name: "should err with bad userID in context",
			userIDInContext: uuid.Nil,
            accessTokenInCtx: testAccessToken,
            pathParams: map[string]string{"item-id": "12345"},
            mockDb: &mockDatabaseService{
                DeleteItemFunc: func(ctx context.Context, arg database.DeleteItemParams) error {
                    return nil 
                },  
            },
            mockPlaidService: &mockPlaidService{
                RemoveItemFunc: func(ctx context.Context, accessToken string) error {
                    return nil
                },
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad userID in context",
        },
        {
            name: "should err on deleting item",
			userIDInContext: testUserID,
            accessTokenInCtx: testAccessToken,
            pathParams: map[string]string{"item-id": "12345"},
            mockDb: &mockDatabaseService{
                DeleteItemFunc: func(ctx context.Context, arg database.DeleteItemParams) error {
                    return fmt.Errorf("mock error")
                },  
            },
            mockPlaidService: &mockPlaidService{
                RemoveItemFunc: func(ctx context.Context, accessToken string) error {
                    return nil
                },
            },
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
        {
            name: "should err on deleting item from plaid",
			userIDInContext: testUserID,
            accessTokenInCtx: testAccessToken,
            pathParams: map[string]string{"item-id": "12345"},
            mockDb: &mockDatabaseService{
                DeleteItemFunc: func(ctx context.Context, arg database.DeleteItemParams) error {
                    return nil
                },  
            },
            mockPlaidService: &mockPlaidService{
                RemoveItemFunc: func(ctx context.Context, accessToken string) error {
                    return fmt.Errorf("mock error")
                },
            },
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Service error",
        },
    }

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/items/%s", tt.pathParams["item-id"]), bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            
            rctx := chi.NewRouteContext()
            for key, value := range tt.pathParams {
                rctx.URLParams.Add(key, value)
            }

            ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
            ctx = context.WithValue(ctx,handlers.GetUserIDContextKey(), tt.userIDInContext)
            ctx = context.WithValue(ctx,handlers.GetAccessTokenKey(), tt.accessTokenInCtx)

            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
                Logger: kitlog.NewNopLogger(),
                PService: tt.mockPlaidService,
            }

            mockApp.HandlerDeleteItem(rr, req)

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

func TestHandlerGetAccountsForItem(t *testing.T) {
    tests := []struct {
        name            string
		userIDInContext uuid.UUID
        pathParams      map[string]string
        requestBody     string
        mockDb          *mockDatabaseService
		mockAuth 		*mockAuthService
        expectedStatus  int 
        expectedBody    string  
    }{
		{
            name: "should successfully get account records for item",
            pathParams: map[string]string{"item-id": "12345"},
            mockDb: &mockDatabaseService{
                GetAccountsForItemFunc: func(ctx context.Context, itemID string) ([]database.Account, error) {
                    return []database.Account{{ID: testAccountID, Name: testAccountName}}, nil
                }, 
            },
            expectedStatus: http.StatusOK,
            expectedBody: testAccountID,
        },
        {
            name: "should err with no accounts found",
            pathParams: map[string]string{"item-id": "12345"},
            mockDb: &mockDatabaseService{
                GetAccountsForItemFunc: func(ctx context.Context, itemID string) ([]database.Account, error) {
                    return []database.Account{}, sql.ErrNoRows
                }, 
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: "No accounts found for item",
        },
        {
            name: "should err on getting accounts for item",
            pathParams: map[string]string{"item-id": "12345"},
            mockDb: &mockDatabaseService{
                GetAccountsForItemFunc: func(ctx context.Context, itemID string) ([]database.Account, error) {
                    return []database.Account{}, fmt.Errorf("mock error")
                }, 
            },
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
    }

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", fmt.Sprintf("/api/items/%s/accounts", tt.pathParams["item-id"]), bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            
            rctx := chi.NewRouteContext()
            for key, value := range tt.pathParams {
                rctx.URLParams.Add(key, value)
            }

            ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
            ctx = context.WithValue(ctx,handlers.GetUserIDContextKey(), tt.userIDInContext)

            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
                Logger: kitlog.NewNopLogger(),
            }

            mockApp.HandlerGetAccountsForItem(rr, req)

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
