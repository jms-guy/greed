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
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/server/handlers"
)

func TestHandlerGetCurrentUser(t *testing.T) {
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
			name:            "should successfully return user record",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:        testUserID,
						Name:      testName,
						Email:     testEmail,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   testName,
		},
		{
			name:            "should err with bad user ID",
			userIDInContext: uuid.Nil,
			mockDb:          &mockDatabaseService{},
			expectedStatus:  http.StatusBadRequest,
			expectedBody:    "Bad userID in context",
		},
		{
			name:            "should err with no user found",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{}, sql.ErrNoRows
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "User not found in database",
		},
		{
			name:            "should err on getting user",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{}, fmt.Errorf("mock error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/users/me", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), handlers.GetUserIDContextKey(), tt.userIDInContext)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			mockApp := &handlers.AppServer{
				Db:     tt.mockDb,
				Logger: kitlog.NewNopLogger(),
			}

			mockApp.HandlerGetCurrentUser(rr, req)

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

func TestHandlerDeleteUser(t *testing.T) {
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
			name:            "should successfully delete user record",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:        testUserID,
						Name:      testName,
						Email:     testEmail,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil
				},
				DeleteUserFunc: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "User deleted successfully",
		},
		{
			name:            "should err with bad user ID",
			userIDInContext: uuid.Nil,
			mockDb:          &mockDatabaseService{},
			expectedStatus:  http.StatusBadRequest,
			expectedBody:    "Bad userID in context",
		},
		{
			name:            "should err with no user found",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{}, sql.ErrNoRows
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "User not found in database",
		},
		{
			name:            "should err on getting user",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{}, fmt.Errorf("mock error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
		{
			name:            "should err on deleting user",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:        testUserID,
						Name:      testName,
						Email:     testEmail,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil
				},
				DeleteUserFunc: func(ctx context.Context, id uuid.UUID) error {
					return fmt.Errorf("mock error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/users/me", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), handlers.GetUserIDContextKey(), tt.userIDInContext)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			mockApp := &handlers.AppServer{
				Db:     tt.mockDb,
				Logger: kitlog.NewNopLogger(),
			}

			mockApp.HandlerDeleteUser(rr, req)

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

func TestHandlerUpdatePassword(t *testing.T) {
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
			name:            "should successfully update password",
			userIDInContext: testUserID,
			requestBody:     `{"new_password": "new_pass", "code": "VALID_CODE"}`,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:         testUserID,
						Name:       testName,
						Email:      testEmail,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
						IsVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil
				},
				GetVerificationRecordFunc: func(ctx context.Context, verificationCode string) (database.VerificationRecord, error) {
					return database.VerificationRecord{
						UserID:           testUserID,
						VerificationCode: "VALID_CODE",
						ExpiryTime:       time.Now().Add(time.Hour).Truncate(time.Second),
					}, nil
				},
				DeleteVerificationRecordFunc: func(ctx context.Context, verificationCode string) error {
					return nil
				},
				UpdatePasswordFunc: func(ctx context.Context, arg database.UpdatePasswordParams) error {
					return nil
				},
			},
			mockAuth: &mockAuthService{
				HashPasswordFunc: func(password string) (string, error) {
					return "new_pass_hash", nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "new_pass_hash",
		},
		{
			name:            "should err with bad userID",
			userIDInContext: uuid.Nil,
			mockDb:          &mockDatabaseService{},
			mockAuth:        &mockAuthService{},
			expectedStatus:  http.StatusBadRequest,
			expectedBody:    "Bad userID in context",
		},
		{
			name:            "should err getting user (sql.ErrNoRows)",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{}, sql.ErrNoRows
				},
			},
			mockAuth:       &mockAuthService{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "No user record found",
		},
		{
			name:            "should err on getting user",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{}, fmt.Errorf("mock error")
				},
			},
			mockAuth:       &mockAuthService{},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
		{
			name:            "should err, user's email not verified",
			userIDInContext: testUserID,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:         testUserID,
						Name:       testName,
						Email:      testEmail,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
						IsVerified: sql.NullBool{Bool: false, Valid: true},
					}, nil
				},
			},
			mockAuth:       &mockAuthService{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "User's email is not verified",
		},
		{
			name:            "should err on decoding request",
			userIDInContext: testUserID,
			requestBody:     "",
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:         testUserID,
						Name:       testName,
						Email:      testEmail,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
						IsVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad request",
		},
		{
			name:            "should err, no verification record found",
			userIDInContext: testUserID,
			requestBody:     `{"new_password": "new_pass", "code": "VALID_CODE"}`,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:         testUserID,
						Name:       testName,
						Email:      testEmail,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
						IsVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil
				},
				GetVerificationRecordFunc: func(ctx context.Context, verificationCode string) (database.VerificationRecord, error) {
					return database.VerificationRecord{}, sql.ErrNoRows
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "No verification record found",
		},
		{
			name:            "should err, on getting verification record",
			userIDInContext: testUserID,
			requestBody:     `{"new_password": "new_pass", "code": "VALID_CODE"}`,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:         testUserID,
						Name:       testName,
						Email:      testEmail,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
						IsVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil
				},
				GetVerificationRecordFunc: func(ctx context.Context, verificationCode string) (database.VerificationRecord, error) {
					return database.VerificationRecord{}, fmt.Errorf("mock error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
		{
			name:            "should err, on deleting expired verification record",
			userIDInContext: testUserID,
			requestBody:     `{"new_password": "new_pass", "code": "VALID_CODE"}`,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:         testUserID,
						Name:       testName,
						Email:      testEmail,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
						IsVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil
				},
				GetVerificationRecordFunc: func(ctx context.Context, verificationCode string) (database.VerificationRecord, error) {
					return database.VerificationRecord{
						UserID:           testUserID,
						VerificationCode: "VALID_CODE",
						ExpiryTime:       time.Now().Add(-1 * time.Hour),
					}, nil
				},
				DeleteVerificationRecordFunc: func(ctx context.Context, verificationCode string) error {
					return fmt.Errorf("mock error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
		{
			name:            "should err, expired verification record",
			userIDInContext: testUserID,
			requestBody:     `{"new_password": "new_pass", "code": "VALID_CODE"}`,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:         testUserID,
						Name:       testName,
						Email:      testEmail,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
						IsVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil
				},
				GetVerificationRecordFunc: func(ctx context.Context, verificationCode string) (database.VerificationRecord, error) {
					return database.VerificationRecord{
						UserID:           testUserID,
						VerificationCode: "VALID_CODE",
						ExpiryTime:       time.Now().Add(-1 * time.Hour),
					}, nil
				},
				DeleteVerificationRecordFunc: func(ctx context.Context, verificationCode string) error {
					return nil
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Verification record is expired",
		},
		{
			name:            "should err, on hashing new password",
			userIDInContext: testUserID,
			requestBody:     `{"new_password": "new_pass", "code": "VALID_CODE"}`,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:         testUserID,
						Name:       testName,
						Email:      testEmail,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
						IsVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil
				},
				GetVerificationRecordFunc: func(ctx context.Context, verificationCode string) (database.VerificationRecord, error) {
					return database.VerificationRecord{
						UserID:           testUserID,
						VerificationCode: "VALID_CODE",
						ExpiryTime:       time.Now().Add(time.Hour),
					}, nil
				},
				DeleteVerificationRecordFunc: func(ctx context.Context, verificationCode string) error {
					return nil
				},
			},
			mockAuth: &mockAuthService{
				HashPasswordFunc: func(password string) (string, error) {
					return "", fmt.Errorf("mock error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Error hashing password",
		},
		{
			name:            "should err, on updating password",
			userIDInContext: testUserID,
			requestBody:     `{"new_password": "new_pass", "code": "VALID_CODE"}`,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:         testUserID,
						Name:       testName,
						Email:      testEmail,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
						IsVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil
				},
				GetVerificationRecordFunc: func(ctx context.Context, verificationCode string) (database.VerificationRecord, error) {
					return database.VerificationRecord{
						UserID:           testUserID,
						VerificationCode: "VALID_CODE",
						ExpiryTime:       time.Now().Add(time.Hour),
					}, nil
				},
				DeleteVerificationRecordFunc: func(ctx context.Context, verificationCode string) error {
					return nil
				},
				UpdatePasswordFunc: func(ctx context.Context, arg database.UpdatePasswordParams) error {
					return fmt.Errorf("mock error")
				},
			},
			mockAuth: &mockAuthService{
				HashPasswordFunc: func(password string) (string, error) {
					return "new_pass_hash", nil
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
		{
			name:            "should err, on deleting verification record",
			userIDInContext: testUserID,
			requestBody:     `{"new_password": "new_pass", "code": "VALID_CODE"}`,
			mockDb: &mockDatabaseService{
				GetUserFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
					return database.User{
						ID:         testUserID,
						Name:       testName,
						Email:      testEmail,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
						IsVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil
				},
				GetVerificationRecordFunc: func(ctx context.Context, verificationCode string) (database.VerificationRecord, error) {
					return database.VerificationRecord{
						UserID:           testUserID,
						VerificationCode: "VALID_CODE",
						ExpiryTime:       time.Now().Add(time.Hour),
					}, nil
				},
				DeleteVerificationRecordFunc: func(ctx context.Context, verificationCode string) error {
					return fmt.Errorf("mock error")
				},
				UpdatePasswordFunc: func(ctx context.Context, arg database.UpdatePasswordParams) error {
					return nil
				},
			},
			mockAuth: &mockAuthService{
				HashPasswordFunc: func(password string) (string, error) {
					return "new_pass_hash", nil
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("PUT", "/api/users/update-password", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), handlers.GetUserIDContextKey(), tt.userIDInContext)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			mockApp := &handlers.AppServer{
				Db:     tt.mockDb,
				Auth:   tt.mockAuth,
				Logger: kitlog.NewNopLogger(),
			}

			mockApp.HandlerUpdatePassword(rr, req)

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
