package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/server/handlers"
	"github.com/jms-guy/greed/models"
)

func TestHandlerCreateUser(t *testing.T) {
	t.Run("should create user successfully with 201 status and correct body", func(t *testing.T) {
        expectedName := "testuser"
        expectedEmail := "test@example.com"
        plainTextPassword := "password123" 
        mockHashedPassword := "password123_hashed"

        reqBody := `{"name": "` + expectedName + `", "email": "` + expectedEmail + `", "password": "` + plainTextPassword + `"}`
        req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockUserID := uuid.New()
        mockCreatedAt := time.Now().Truncate(time.Second) 
        mockUpdatedAt := time.Now().Truncate(time.Second) 

        mockDb := &mockDatabaseService{
            GetUserByNameFunc: func(ctx context.Context, name string) (database.User, error) {
                return database.User{}, sql.ErrNoRows 
            },
            CreateUserFunc: func(ctx context.Context, params database.CreateUserParams) (database.User, error) {
                return database.User{
                    ID: mockUserID,
                    Name: params.Name,
                    HashedPassword: mockHashedPassword,
                    Email: params.Email,
                    CreatedAt: mockCreatedAt,
                    UpdatedAt: mockUpdatedAt,
                }, nil
            },
        }

        mockAuth := &mockAuthService{
            EmailValidationFunc: func(email string) bool {
                return true
            },
            HashPasswordFunc: func(password string) (string, error) {
                return mockHashedPassword, nil 
            },
        }

        mockApp := &handlers.AppServer{
            Db: mockDb,
            Auth: mockAuth,
            Logger: kitlog.NewNopLogger(), 
        }

        mockApp.HandlerCreateUser(rr, req)

        // --- Assertions ---
        if status := rr.Code; status != http.StatusCreated {
            t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusCreated, rr.Body.String())
            return 
        }

        var actualUser models.User
        err := json.Unmarshal(rr.Body.Bytes(), &actualUser)
        if err != nil {
            t.Fatalf("could not unmarshal response body: %v", err)
        }

        if actualUser.ID != mockUserID {
            t.Errorf("handler returned wrong user ID: got %v want %v", actualUser.ID, mockUserID)
        }
        if actualUser.Name != expectedName {
            t.Errorf("handler returned wrong user name: got %q want %q", actualUser.Name, expectedName)
        }
        if actualUser.Email != expectedEmail {
            t.Errorf("handler returned wrong user email: got %q want %q", actualUser.Email, expectedEmail)
        }
        if actualUser.HashedPassword != mockHashedPassword {
            t.Errorf("handler returned wrong hashed password: got %q want %q", actualUser.HashedPassword, mockHashedPassword)
        }

        if !actualUser.CreatedAt.Equal(mockCreatedAt) {
            t.Errorf("handler returned wrong CreatedAt: got %v want %v", actualUser.CreatedAt, mockCreatedAt)
        }
        if !actualUser.UpdatedAt.Equal(mockUpdatedAt) {
            t.Errorf("handler returned wrong UpdatedAt: got %v want %v", actualUser.UpdatedAt, mockUpdatedAt)
        }

        if actualUser.CreatedAt.After(actualUser.UpdatedAt) {
            t.Errorf("handler returned CreatedAt (%v) after UpdatedAt (%v)", actualUser.CreatedAt, actualUser.UpdatedAt)
        }
    })
    t.Run("should error, user already exists", func(t *testing.T) {
        expectedName := "testuser"
        expectedEmail := "test@example.com"
        plainTextPassword := "password123" 
        mockHashedPassword := "password123_hashed"
        expectedErrorMessage := "User already exists by that name"

        reqBody := `{"name": "` + expectedName + `", "email": "` + expectedEmail + `", "password": "` + plainTextPassword + `"}`
        req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockUserID := uuid.New()
        mockCreatedAt := time.Now().Truncate(time.Second) 
        mockUpdatedAt := time.Now().Truncate(time.Second) 

        mockDb := &mockDatabaseService{
            GetUserByNameFunc: func(ctx context.Context, name string) (database.User, error) {
                return database.User{
                    ID: mockUserID,
                    Name: expectedName,
                    HashedPassword: mockHashedPassword,
                    Email: expectedEmail,
                    CreatedAt: mockCreatedAt,
                    UpdatedAt: mockUpdatedAt,
                }, nil
            },
            CreateUserFunc: func(ctx context.Context, params database.CreateUserParams) (database.User, error) {
                t.Fatal("CreateUser should not be called when user already exists")
                return database.User{}, nil
            },
        }

        mockAuth := &mockAuthService{
            EmailValidationFunc: func(email string) bool {
                return true
            },
            HashPasswordFunc: func(password string) (string, error) {
                return mockHashedPassword, nil 
            },
        }

        mockApp := &handlers.AppServer{
            Db: mockDb,
            Auth: mockAuth,
            Logger: kitlog.NewNopLogger(), 
        }

        mockApp.HandlerCreateUser(rr, req)

        // --- Assertions ---
        if status := rr.Code; status != http.StatusBadRequest {
            t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusBadRequest, rr.Body.String())
            return 
        }

        var actualErrorResponse errorResponse
        err := json.Unmarshal(rr.Body.Bytes(), &errorResponse{})
        if err != nil {
            t.Fatalf("could not unmarshal error response body: %v", err)
        }

        if actualErrorResponse.Error != expectedErrorMessage {
            t.Errorf("handler returned wrong error message: got %q want %q", actualErrorResponse.Error, expectedErrorMessage)
        }
    })
}   