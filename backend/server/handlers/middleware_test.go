package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	kitlog "github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/server/handlers"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		tokenInContext  string
		requestBody     string
		mockDb          *mockDatabaseService
		mockAuth        *mockAuthService
		expectedStatus  int
		expectedContext uuid.UUID
	}{
		{
			name:           "should successfully place test userID in context",
			tokenInContext: "testJWT",
			mockDb:         &mockDatabaseService{},
			mockAuth: &mockAuthService{
				GetBearerTokenFunc: func(headers http.Header) (string, error) {
					return "testJWT", nil
				},
				ValidateJWTFunc: func(cfg *config.Config, tokenString string) (uuid.UUID, error) {
					return testUserID, nil
				},
			},
			expectedStatus:  http.StatusOK,
			expectedContext: testUserID,
		},
		{
			name:           "should err with bad token",
			tokenInContext: "testJWT",
			mockDb:         &mockDatabaseService{},
			mockAuth: &mockAuthService{
				GetBearerTokenFunc: func(headers http.Header) (string, error) {
					return "testJWT", fmt.Errorf("mock error")
				},
				ValidateJWTFunc: func(cfg *config.Config, tokenString string) (uuid.UUID, error) {
					return testUserID, nil
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should err with invalid JWT",
			tokenInContext: "testJWT",
			mockDb:         &mockDatabaseService{},
			mockAuth: &mockAuthService{
				GetBearerTokenFunc: func(headers http.Header) (string, error) {
					return "testJWT", nil
				},
				ValidateJWTFunc: func(cfg *config.Config, tokenString string) (uuid.UUID, error) {
					return testUserID, fmt.Errorf("mock error")
				},
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Authorization", "Bearer "+tt.tokenInContext)

			rr := httptest.NewRecorder()

			mockApp := &handlers.AppServer{
				Db:     tt.mockDb,
				Auth:   tt.mockAuth,
				Config: &config.Config{},
				Logger: kitlog.NewNopLogger(),
			}

			var userIDFromContext uuid.UUID

			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if val := r.Context().Value(handlers.GetUserIDContextKey()); val != nil {
					id, ok := val.(uuid.UUID)
					if ok && id != uuid.Nil {
						userIDFromContext = id
					}
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			handler := mockApp.AuthMiddleware(dummyHandler)
			handler.ServeHTTP(rr, req)

			// --- Assertions ---
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, tt.expectedStatus, rr.Body.String())
			}

			if userIDFromContext != tt.expectedContext {
				t.Errorf("handler did not set expected userID in context: got %v want %v", userIDFromContext, tt.expectedContext)
			}

		})
	}
}

func TestAccessTokenMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		userIDInContext uuid.UUID
		requestBody     string
		mockDb          *mockDatabaseService
		mockAuth        *mockAuthService
		mockEncryptor   *mockEncryptor
		expectedStatus  int
		expectedContext string
		expectedBody    string
	}{
		{
			name:            "should successfully place access token in context",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetAccessTokenFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{UserID: testUserID, AccessToken: testAccessToken}, nil
				},
			},
			mockAuth: &mockAuthService{},
			mockEncryptor: &mockEncryptor{
				DecryptAccessTokenFunc: func(ciphertext string, keyString string) ([]byte, error) {
					return []byte(testAccessToken), nil
				},
			},
			expectedStatus:  http.StatusOK,
			expectedContext: testAccessToken,
		},
		{
			name:            "should err with bad userID in context",
			userIDInContext: uuid.Nil,
			mockDb: &mockDatabaseService{
				GetAccessTokenFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{UserID: testUserID, AccessToken: testAccessToken}, nil
				},
			},
			mockAuth: &mockAuthService{},
			mockEncryptor: &mockEncryptor{
				DecryptAccessTokenFunc: func(ciphertext string, keyString string) ([]byte, error) {
					return []byte(testAccessToken), nil
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad userID in context",
		},
		{
			name:            "should err with no access token found for user",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetAccessTokenFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{UserID: testUserID, AccessToken: testAccessToken}, sql.ErrNoRows
				},
			},
			mockAuth: &mockAuthService{},
			mockEncryptor: &mockEncryptor{
				DecryptAccessTokenFunc: func(ciphertext string, keyString string) ([]byte, error) {
					return []byte(testAccessToken), nil
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "No item found for user",
		},
		{
			name:            "should err on getting access token for user",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetAccessTokenFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{UserID: testUserID, AccessToken: testAccessToken}, fmt.Errorf("mock error")
				},
			},
			mockAuth: &mockAuthService{},
			mockEncryptor: &mockEncryptor{
				DecryptAccessTokenFunc: func(ciphertext string, keyString string) ([]byte, error) {
					return []byte(testAccessToken), nil
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
		{
			name:            "should err with mismatching userIDs",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetAccessTokenFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{UserID: uuid.New(), AccessToken: testAccessToken}, nil
				},
			},
			mockAuth: &mockAuthService{},
			mockEncryptor: &mockEncryptor{
				DecryptAccessTokenFunc: func(ciphertext string, keyString string) ([]byte, error) {
					return []byte(testAccessToken), nil
				},
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "UserID does not match item's database record",
		},
		{
			name:            "should err on decrypting access token",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetAccessTokenFunc: func(ctx context.Context, id string) (database.PlaidItem, error) {
					return database.PlaidItem{UserID: testUserID, AccessToken: testAccessToken}, nil
				},
			},
			mockAuth: &mockAuthService{},
			mockEncryptor: &mockEncryptor{
				DecryptAccessTokenFunc: func(ciphertext string, keyString string) ([]byte, error) {
					return []byte(testAccessToken), fmt.Errorf("mock error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Error decrypting access token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.requestBody))
			ctx := context.WithValue(req.Context(), handlers.GetUserIDContextKey(), tt.userIDInContext)

			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			mockApp := &handlers.AppServer{
				Db:        tt.mockDb,
				Auth:      tt.mockAuth,
				Config:    &config.Config{},
				Logger:    kitlog.NewNopLogger(),
				Encryptor: tt.mockEncryptor,
			}

			var tokenFromContext string

			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if val := r.Context().Value(handlers.GetAccessTokenKey()); val != nil {
					token, ok := val.(string)
					if ok {
						tokenFromContext = token
					}
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			handler := mockApp.AccessTokenMiddleware(dummyHandler)
			handler.ServeHTTP(rr, req)

			// --- Assertions ---
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, tt.expectedStatus, rr.Body.String())
			}

			if tokenFromContext != tt.expectedContext {
				t.Errorf("handler did not set expected token in context: got %v want %v", tokenFromContext, tt.expectedContext)
			}

			if !strings.Contains(rr.Body.String(), tt.expectedBody) {
				t.Errorf("handler returned unexpected body: got %s want body to contain %s", rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestMemberMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		userIDInContext uuid.UUID
		requestBody     string
		mockDb          *mockDatabaseService
		mockAuth        *mockAuthService
		expectedStatus  int
		expectedBody    string
	}{
		{
			name:            "should successfully call next handler when user is member",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{ID: testUserID, IsMember: true, FreeCalls: 5}, nil
				},
				UpdateFreeCallsFunc: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
			},
			mockAuth:       &mockAuthService{},
			expectedStatus: http.StatusOK,
		},
		{
			name:            "should successfully call next handler when user is not member but has free calls",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{ID: testUserID, IsMember: false, FreeCalls: 5}, nil
				},
				UpdateFreeCallsFunc: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
			},
			mockAuth:       &mockAuthService{},
			expectedStatus: http.StatusOK,
		},
		{
			name:            "should err when user is not member and has no free calls",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{ID: testUserID, IsMember: false, FreeCalls: 0}, nil
				},
				UpdateFreeCallsFunc: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
			},
			mockAuth:       &mockAuthService{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad request",
		},
		{
			name:            "should err on updating user's free calls",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{ID: testUserID, IsMember: false, FreeCalls: 2}, nil
				},
				UpdateFreeCallsFunc: func(ctx context.Context, id uuid.UUID) error {
					return fmt.Errorf("mock error")
				},
			},
			mockAuth:       &mockAuthService{},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
		{
			name:            "should err on getting user",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{ID: testUserID, IsMember: false, FreeCalls: 0}, fmt.Errorf("mock error")
				},
				UpdateFreeCallsFunc: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
			},
			mockAuth:       &mockAuthService{},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
		{
			name:            "should err with bad userID in context",
			userIDInContext: uuid.Nil,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{ID: testUserID, IsMember: false, FreeCalls: 0}, nil
				},
				UpdateFreeCallsFunc: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
			},
			mockAuth:       &mockAuthService{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad userID in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.requestBody))
			ctx := context.WithValue(req.Context(), handlers.GetUserIDContextKey(), tt.userIDInContext)

			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			mockApp := &handlers.AppServer{
				Db:     tt.mockDb,
				Auth:   tt.mockAuth,
				Config: &config.Config{},
				Logger: kitlog.NewNopLogger(),
			}

			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			handler := mockApp.MemberMiddleware(dummyHandler)
			handler.ServeHTTP(rr, req)

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

func TestAccountMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		userIDInContext uuid.UUID
		pathParams      map[string]string
		requestBody     string
		mockDb          *mockDatabaseService
		mockAuth        *mockAuthService
		expectedStatus  int
		expectedContext database.Account
		expectedBody    string
	}{
		{
			name:            "should successfully place account in context",
			userIDInContext: testUserID,
			pathParams:      map[string]string{"accountid": testAccountID},
			mockDb: &mockDatabaseService{
				GetAccountByIdFunc: func(ctx context.Context, arg database.GetAccountByIdParams) (database.Account, error) {
					return testAccount, nil
				},
			},
			mockAuth:        &mockAuthService{},
			expectedStatus:  http.StatusOK,
			expectedContext: testAccount,
		},
		{
			name:            "should err with bad userID in context",
			userIDInContext: uuid.Nil,
			pathParams:      map[string]string{"accountid": testAccountID},
			mockDb: &mockDatabaseService{
				GetAccountByIdFunc: func(ctx context.Context, arg database.GetAccountByIdParams) (database.Account, error) {
					return testAccount, nil
				},
			},
			mockAuth:       &mockAuthService{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad userID in context",
		},
		{
			name:            "should err with no account found",
			userIDInContext: testUserID,
			pathParams:      map[string]string{"accountid": testAccountID},
			mockDb: &mockDatabaseService{
				GetAccountByIdFunc: func(ctx context.Context, arg database.GetAccountByIdParams) (database.Account, error) {
					return testAccount, sql.ErrNoRows
				},
			},
			mockAuth:       &mockAuthService{},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Account not found",
		},
		{
			name:            "should err on getting account from database",
			userIDInContext: testUserID,
			pathParams:      map[string]string{"accountid": testAccountID},
			mockDb: &mockDatabaseService{
				GetAccountByIdFunc: func(ctx context.Context, arg database.GetAccountByIdParams) (database.Account, error) {
					return testAccount, fmt.Errorf("mock error")
				},
			},
			mockAuth:       &mockAuthService{},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.requestBody))
			ctx := context.WithValue(req.Context(), handlers.GetUserIDContextKey(), tt.userIDInContext)

			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			mockApp := &handlers.AppServer{
				Db:     tt.mockDb,
				Auth:   tt.mockAuth,
				Config: &config.Config{},
				Logger: kitlog.NewNopLogger(),
			}

			var accountFromContext database.Account

			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if val := r.Context().Value(handlers.GetAccountKey()); val != nil {
					account, ok := val.(database.Account)
					if ok {
						accountFromContext = account
					}
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			handler := mockApp.AccountMiddleware(dummyHandler)
			handler.ServeHTTP(rr, req)

			// --- Assertions ---
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, tt.expectedStatus, rr.Body.String())
			}

			if accountFromContext != tt.expectedContext {
				t.Errorf("handler did not set expected token in context: got %v want %v", accountFromContext, tt.expectedContext)
			}

			if !strings.Contains(rr.Body.String(), tt.expectedBody) {
				t.Errorf("handler returned unexpected body: got %s want body to contain %s", rr.Body.String(), tt.expectedBody)
			}
		})
	}
}
