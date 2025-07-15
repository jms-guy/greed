package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	kitlog "github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/config"
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

        var actualErrorResponse handlers.ErrorResponse
        err := json.Unmarshal(rr.Body.Bytes(), &actualErrorResponse)
        if err != nil {
            t.Fatalf("could not unmarshal error response body: %v", err)
        }

        if actualErrorResponse.Error != expectedErrorMessage {
            t.Errorf("handler returned wrong error message: got %q want %q", actualErrorResponse.Error, expectedErrorMessage)
        }
    })
    t.Run("should error, invalid email", func(t *testing.T) {
        expectedName := "testuser"
        expectedEmail := "test@example"
        plainTextPassword := "password123" 
        mockHashedPassword := "password123_hashed"
        expectedErrorMessage := "Email provided is invalid"

        reqBody := `{"name": "` + expectedName + `", "email": "` + expectedEmail + `", "password": "` + plainTextPassword + `"}`
        req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockDb := &mockDatabaseService{
            GetUserByNameFunc: func(ctx context.Context, name string) (database.User, error) {
                return database.User{}, sql.ErrNoRows 
            },
            CreateUserFunc: func(ctx context.Context, params database.CreateUserParams) (database.User, error) {
                t.Fatal("CreateUser should not be called when email is invalid")
                return database.User{}, nil
            },
        }

        mockAuth := &mockAuthService{
            EmailValidationFunc: func(email string) bool {
                return false
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

        var actualErrorResponse handlers.ErrorResponse
        err := json.Unmarshal(rr.Body.Bytes(), &actualErrorResponse)
        if err != nil {
            t.Fatalf("could not unmarshal error response body: %v", err)
        }

        if actualErrorResponse.Error != expectedErrorMessage {
            t.Errorf("handler returned wrong error message: got %q want %q", actualErrorResponse.Error, expectedErrorMessage)
        }
    })
    t.Run("should error, from getting user", func(t *testing.T) {
        expectedName := "testuser"
        expectedEmail := "test@example.com"
        plainTextPassword := "password123" 
        mockHashedPassword := "password123_hashed"
        expectedErrorMessage := "Database error"

        reqBody := `{"name": "` + expectedName + `", "email": "` + expectedEmail + `", "password": "` + plainTextPassword + `"}`
        req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockDb := &mockDatabaseService{
            GetUserByNameFunc: func(ctx context.Context, name string) (database.User, error) {
                return database.User{}, fmt.Errorf("mock error: getting user error")
            },
            CreateUserFunc: func(ctx context.Context, params database.CreateUserParams) (database.User, error) {
                t.Fatal("Should not be called when err's")
                return database.User{}, nil
            },
        }

        mockAuth := &mockAuthService{
            EmailValidationFunc: func(email string) bool {
                t.Fatal("Should not be called when err's")
                return true
            },
            HashPasswordFunc: func(password string) (string, error) {
                t.Fatal("Should not be called when err's")
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
        if status := rr.Code; status != http.StatusInternalServerError {
            t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusInternalServerError, rr.Body.String())
            return 
        }

        var actualErrorResponse handlers.ErrorResponse
        err := json.Unmarshal(rr.Body.Bytes(), &actualErrorResponse)
        if err != nil {
            t.Fatalf("could not unmarshal error response body: %v", err)
        }

        if actualErrorResponse.Error != expectedErrorMessage {
            t.Errorf("handler returned wrong error message: got %q want %q", actualErrorResponse.Error, expectedErrorMessage)
        }
    })
    t.Run("should error, from creating user", func(t *testing.T) {
        expectedName := "testuser"
        expectedEmail := "test@example.com"
        plainTextPassword := "password123" 
        mockHashedPassword := "password123_hashed"
        expectedErrorMessage := "Database error"

        reqBody := `{"name": "` + expectedName + `", "email": "` + expectedEmail + `", "password": "` + plainTextPassword + `"}`
        req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockDb := &mockDatabaseService{
            GetUserByNameFunc: func(ctx context.Context, name string) (database.User, error) {
                return database.User{}, sql.ErrNoRows
            },
            CreateUserFunc: func(ctx context.Context, params database.CreateUserParams) (database.User, error) {
                return database.User{}, fmt.Errorf("mock error: creating user error")
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
        if status := rr.Code; status != http.StatusInternalServerError {
            t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusInternalServerError, rr.Body.String())
            return 
        }

        var actualErrorResponse handlers.ErrorResponse
        err := json.Unmarshal(rr.Body.Bytes(), &actualErrorResponse)
        if err != nil {
            t.Fatalf("could not unmarshal error response body: %v", err)
        }

        if actualErrorResponse.Error != expectedErrorMessage {
            t.Errorf("handler returned wrong error message: got %q want %q", actualErrorResponse.Error, expectedErrorMessage)
        }
    })
}   

func TestHandlerUserLogin(t *testing.T) {
    t.Run("should login successfully with 200 response code and credentials struct", func(t *testing.T) {
        expectedName := "testuser"
        expectedEmail := "test@example.com"
        plainTextPassword := "password123" 
        mockHashedPassword := "password123_hashed"

        reqBody := `{"name": "` + expectedName + `", "email": "` + expectedEmail + `", "password": "` + plainTextPassword + `"}`
        req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockUserID := uuid.New()
        mockCreatedAt := time.Now().Truncate(time.Second) 
        mockUpdatedAt := time.Now().Truncate(time.Second) 

        mockDelegationID := uuid.New()
        mockExpiresAt := time.Now().Add(1 * time.Hour).Truncate(time.Second)
        mockRevokedAt := sql.NullTime{}

        expectedJWT := "JWTToken"
        expectedRefresh := "refreshToken"

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
            CreateDelegationFunc: func(ctx context.Context, arg database.CreateDelegationParams) (database.Delegation, error) {
                return database.Delegation{
                    ID: mockDelegationID,
                    UserID: mockUserID,
                    CreatedAt: mockCreatedAt,
                    ExpiresAt: mockExpiresAt,
                    RevokedAt: mockRevokedAt,
                    LastUsed: mockCreatedAt,
                }, nil
            },
        }

        mockAuth := &mockAuthService{
            ValidatePasswordHashFunc: func(hash string, password string) error {
                return nil
            },
            MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
                return "JWTToken", nil
            },
            MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error)  {
                return "refreshToken", nil 
            },
        }

        mockApp := &handlers.AppServer{
            Db: mockDb,
            Auth: mockAuth,
            Config: &config.Config{
                RefreshExpiration: "3600",
            },
            Logger: kitlog.NewNopLogger(), 
        }

        mockApp.HandlerUserLogin(rr, req)

        // --- Assertions ---
        if status := rr.Code; status != http.StatusOK {
            t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
            return 
        }

        var actualCredentials models.Credentials
        err := json.Unmarshal(rr.Body.Bytes(), &actualCredentials)
        if err != nil {
            t.Fatalf("could not unmarshal error response body: %v", err)
        }

        if actualCredentials.User.ID != mockUserID {
            t.Errorf("handler returned wrong user ID: got %v want %v", actualCredentials.User.ID, mockUserID)
        }
        if actualCredentials.User.Email != expectedEmail {
            t.Errorf("handler returned wrong email: got %v want %v", actualCredentials.User.Email, expectedEmail)
        }
        if actualCredentials.User.Name != expectedName {
            t.Errorf("handler returned wrong user name: got %v want %v", actualCredentials.User.Name, expectedName)
        }
        if actualCredentials.RefreshToken != expectedRefresh {
            t.Errorf("handler returned wrong refresh token: got %v want %v", actualCredentials.RefreshToken, expectedRefresh)
        }
        if actualCredentials.AccessToken != expectedJWT {
            t.Errorf("handler returned wrong JWT: got %v want %v", actualCredentials.AccessToken, expectedJWT)
        }
    })
    t.Run("should error, incorrect password", func(t *testing.T) {
        expectedName := "testuser"
        expectedEmail := "test@example.com"
        plainTextPassword := "wrongpassword" 
        mockHashedPassword := "password123_hashed"

        reqBody := `{"name": "` + expectedName + `", "email": "` + expectedEmail + `", "password": "` + plainTextPassword + `"}`
        req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockUserID := uuid.New()
        mockCreatedAt := time.Now().Truncate(time.Second) 
        mockUpdatedAt := time.Now().Truncate(time.Second) 

         expectedErrorMessage := "Password is incorrect"

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
            CreateDelegationFunc: func(ctx context.Context, arg database.CreateDelegationParams) (database.Delegation, error) {
                t.Fatalf("should not be called if password is wrong")
                return database.Delegation{}, nil
            },
        }

        mockAuth := &mockAuthService{
            ValidatePasswordHashFunc: func(hash string, password string) error {
                return fmt.Errorf("mock error: passwords do not match")
            },
            MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
                t.Fatal("MakeRefreshToken should not be called if password is wrong")
                return "", nil
            },
            MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error)  {
                return "", nil 
            },
        }

        mockApp := &handlers.AppServer{
            Db: mockDb,
            Auth: mockAuth,
            Config: &config.Config{
                RefreshExpiration: "3600",
            },
            Logger: kitlog.NewNopLogger(), 
        }

        mockApp.HandlerUserLogin(rr, req)

        // --- Assertions ---
        if status := rr.Code; status != http.StatusBadRequest {
            t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusBadRequest, rr.Body.String())
            return 
        }

        var actualErrorResponse handlers.ErrorResponse
        err := json.Unmarshal(rr.Body.Bytes(), &actualErrorResponse)
        if err != nil {
            t.Fatalf("could not unmarshal error response body: %v", err)
        }

        if actualErrorResponse.Error != expectedErrorMessage {
            t.Errorf("handler returned wrong error message: got %q want %q", actualErrorResponse.Error, expectedErrorMessage)
        }
    })
    t.Run("should error, database error getting user record", func(t *testing.T) {
        expectedName := "testuser"
        expectedEmail := "test@example.com"
        plainTextPassword := "wrongpassword" 

        reqBody := `{"name": "` + expectedName + `", "email": "` + expectedEmail + `", "password": "` + plainTextPassword + `"}`
        req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()


        expectedErrorMessage := "Database error"

        mockDb := &mockDatabaseService{
            GetUserByNameFunc: func(ctx context.Context, name string) (database.User, error) {
                return database.User{}, fmt.Errorf("mock error")
            },
            CreateDelegationFunc: func(ctx context.Context, arg database.CreateDelegationParams) (database.Delegation, error) {
                t.Fatalf("should not be called if errs")
                return database.Delegation{}, nil
            },
        }

        mockAuth := &mockAuthService{
            ValidatePasswordHashFunc: func(hash string, password string) error {
                t.Fatalf("should not be called if errs")
                return nil
            },
            MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
                t.Fatalf("should not be called if errs")
                return "", nil
            },
            MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error)  {
                t.Fatalf("should not be called if errs")
                return "", nil 
            },
        }

        mockApp := &handlers.AppServer{
            Db: mockDb,
            Auth: mockAuth,
            Config: &config.Config{
                RefreshExpiration: "3600",
            },
            Logger: kitlog.NewNopLogger(), 
        }

        mockApp.HandlerUserLogin(rr, req)

        // --- Assertions ---
        if status := rr.Code; status != http.StatusInternalServerError {
            t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusInternalServerError, rr.Body.String())
            return 
        }

        var actualErrorResponse handlers.ErrorResponse
        err := json.Unmarshal(rr.Body.Bytes(), &actualErrorResponse)
        if err != nil {
            t.Fatalf("could not unmarshal error response body: %v", err)
        }

        if actualErrorResponse.Error != expectedErrorMessage {
            t.Errorf("handler returned wrong error message: got %q want %q", actualErrorResponse.Error, expectedErrorMessage)
        }
    })
    t.Run("should error, database error creating token delegation", func(t *testing.T) {
        expectedName := "testuser"
        expectedEmail := "test@example.com"
        plainTextPassword := "wrongpassword" 
        mockHashedPassword := "password123_hashed"

        reqBody := `{"name": "` + expectedName + `", "email": "` + expectedEmail + `", "password": "` + plainTextPassword + `"}`
        req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockUserID := uuid.New()
        mockCreatedAt := time.Now().Truncate(time.Second) 
        mockUpdatedAt := time.Now().Truncate(time.Second) 

        expectedErrorMessage := "Database error"

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
            CreateDelegationFunc: func(ctx context.Context, arg database.CreateDelegationParams) (database.Delegation, error) {
                return database.Delegation{}, fmt.Errorf("mock error")
            },
        }

        mockAuth := &mockAuthService{
            ValidatePasswordHashFunc: func(hash string, password string) error {
                return nil
            },
            MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
                t.Fatal("should not be called if errs")
                return "", nil
            },
            MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error)  {
                t.Fatalf("should not be called if errs")
                return "", nil 
            },
        }

        mockApp := &handlers.AppServer{
            Db: mockDb,
            Auth: mockAuth,
            Config: &config.Config{
                RefreshExpiration: "3600",
            },
            Logger: kitlog.NewNopLogger(), 
        }

        mockApp.HandlerUserLogin(rr, req)

        // --- Assertions ---
        if status := rr.Code; status != http.StatusInternalServerError {
            t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusInternalServerError, rr.Body.String())
            return 
        }

        var actualErrorResponse handlers.ErrorResponse
        err := json.Unmarshal(rr.Body.Bytes(), &actualErrorResponse)
        if err != nil {
            t.Fatalf("could not unmarshal error response body: %v", err)
        }

        if actualErrorResponse.Error != expectedErrorMessage {
            t.Errorf("handler returned wrong error message: got %q want %q", actualErrorResponse.Error, expectedErrorMessage)
        }
    })
    t.Run("should error from creating JWT", func(t *testing.T) {
        expectedName := "testuser"
        expectedEmail := "test@example.com"
        plainTextPassword := "password123" 
        mockHashedPassword := "password123_hashed"
        expectedErrorMessage := "Error creating JWT"

        reqBody := `{"name": "` + expectedName + `", "email": "` + expectedEmail + `", "password": "` + plainTextPassword + `"}`
        req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockUserID := uuid.New()
        mockCreatedAt := time.Now().Truncate(time.Second) 
        mockUpdatedAt := time.Now().Truncate(time.Second) 

        mockDelegationID := uuid.New()
        mockExpiresAt := time.Now().Add(1 * time.Hour).Truncate(time.Second)
        mockRevokedAt := sql.NullTime{}

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
            CreateDelegationFunc: func(ctx context.Context, arg database.CreateDelegationParams) (database.Delegation, error) {
                return database.Delegation{
                    ID: mockDelegationID,
                    UserID: mockUserID,
                    CreatedAt: mockCreatedAt,
                    ExpiresAt: mockExpiresAt,
                    RevokedAt: mockRevokedAt,
                    LastUsed: mockCreatedAt,
                }, nil
            },
        }

        mockAuth := &mockAuthService{
            ValidatePasswordHashFunc: func(hash string, password string) error {
                return nil
            },
            MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
                return "", fmt.Errorf("mock error")
            },
            MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error)  {
                t.Fatalf("should not be called if errs")
                return "", nil 
            },
        }

        mockApp := &handlers.AppServer{
            Db: mockDb,
            Auth: mockAuth,
            Config: &config.Config{
                RefreshExpiration: "3600",
            },
            Logger: kitlog.NewNopLogger(), 
        }

        mockApp.HandlerUserLogin(rr, req)

        // --- Assertions ---
        if status := rr.Code; status != http.StatusInternalServerError {
            t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusInternalServerError, rr.Body.String())
            return 
        }

        var actualErrorResponse handlers.ErrorResponse
        err := json.Unmarshal(rr.Body.Bytes(), &actualErrorResponse)
        if err != nil {
            t.Fatalf("could not unmarshal error response body: %v", err)
        }

        if actualErrorResponse.Error != expectedErrorMessage {
            t.Errorf("handler returned wrong error message: got %q want %q", actualErrorResponse.Error, expectedErrorMessage)
        }
    })
    t.Run("should error from creating refresh token", func(t *testing.T) {
        expectedName := "testuser"
        expectedEmail := "test@example.com"
        plainTextPassword := "password123" 
        mockHashedPassword := "password123_hashed"
        expectedErrorMessage := "Error creating refresh token"

        reqBody := `{"name": "` + expectedName + `", "email": "` + expectedEmail + `", "password": "` + plainTextPassword + `"}`
        req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockUserID := uuid.New()
        mockCreatedAt := time.Now().Truncate(time.Second) 
        mockUpdatedAt := time.Now().Truncate(time.Second) 

        mockDelegationID := uuid.New()
        mockExpiresAt := time.Now().Add(1 * time.Hour).Truncate(time.Second)
        mockRevokedAt := sql.NullTime{}

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
            CreateDelegationFunc: func(ctx context.Context, arg database.CreateDelegationParams) (database.Delegation, error) {
                return database.Delegation{
                    ID: mockDelegationID,
                    UserID: mockUserID,
                    CreatedAt: mockCreatedAt,
                    ExpiresAt: mockExpiresAt,
                    RevokedAt: mockRevokedAt,
                    LastUsed: mockCreatedAt,
                }, nil
            },
        }

        mockAuth := &mockAuthService{
            ValidatePasswordHashFunc: func(hash string, password string) error {
                return nil
            },
            MakeJWTFunc: func(cfg *config.Config, userID uuid.UUID) (string, error) {
                return "", nil
            },
            MakeRefreshTokenFunc: func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error)  {
                return "", fmt.Errorf("mock error") 
            },
        }

        mockApp := &handlers.AppServer{
            Db: mockDb,
            Auth: mockAuth,
            Config: &config.Config{
                RefreshExpiration: "3600",
            },
            Logger: kitlog.NewNopLogger(), 
        }

        mockApp.HandlerUserLogin(rr, req)

        // --- Assertions ---
        if status := rr.Code; status != http.StatusInternalServerError {
            t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusInternalServerError, rr.Body.String())
            return 
        }

        var actualErrorResponse handlers.ErrorResponse
        err := json.Unmarshal(rr.Body.Bytes(), &actualErrorResponse)
        if err != nil {
            t.Fatalf("could not unmarshal error response body: %v", err)
        }

        if actualErrorResponse.Error != expectedErrorMessage {
            t.Errorf("handler returned wrong error message: got %q want %q", actualErrorResponse.Error, expectedErrorMessage)
        }
    })
}

func TestHandlerUserLogout(t *testing.T) {
    t.Run("should logout successfully with 200 status code", func(t *testing.T) {
        token := "refreshToken"
        mockHashedToken := "refreshToken_hash"

        mockID := uuid.New()
        
        reqBody := `{"refresh_token": "` + token + `"}`
        req := httptest.NewRequest("POST", "/api/auth/logout", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockDb := &mockDatabaseService{
            GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
                return database.RefreshToken{
                    ID: mockID,
                }, nil
            },
            RevokeDelegationByIDFunc: func(ctx context.Context, id uuid.UUID) error {
                return nil
            },
            ExpireAllDelegationTokensFunc: func(ctx context.Context, delegationID uuid.UUID) error {
                return nil 
            },
        }

        mockAuth := &mockAuthService{
            HashRefreshTokenFunc: func(token string) string {
                return mockHashedToken
            },
        }

        mockApp := &handlers.AppServer{
            Db: mockDb,
            Auth: mockAuth,
            Logger: kitlog.NewNopLogger(), 
        }

        mockApp.HandlerUserLogout(rr, req)

        // --- Assertions ---
        if status := rr.Code; status != http.StatusOK {
             t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
            return
        }
    })
    t.Run("should err for no refresh token", func(t *testing.T) {
        token := ""
        
        reqBody := `{"refresh_token": "` + token + `"}`
        req := httptest.NewRequest("POST", "/api/auth/logout", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockDb := &mockDatabaseService{
            GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
                t.Fatalf("should not be called if errs")
                return database.RefreshToken{}, nil
            },
            RevokeDelegationByIDFunc: func(ctx context.Context, id uuid.UUID) error {
                t.Fatalf("should not be called if errs")
                return nil
            },
            ExpireAllDelegationTokensFunc: func(ctx context.Context, delegationID uuid.UUID) error {
                t.Fatalf("should not be called if errs")
                return nil 
            },
        }

        mockAuth := &mockAuthService{
            HashRefreshTokenFunc: func(token string) string {
                return ""
            },
        }

        mockApp := &handlers.AppServer{
            Db: mockDb,
            Auth: mockAuth,
            Logger: kitlog.NewNopLogger(), 
        }

        mockApp.HandlerUserLogout(rr, req)

        // --- Assertions ---
        if status := rr.Code; status != http.StatusBadRequest {
             t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusBadRequest, rr.Body.String())
            return
        }
    })
    t.Run("should error with invalid refresh token", func(t *testing.T) {
        token := "refreshTokenToken"
        mockHashedToken := "refreshToken_hash"
        expectedErrorMessage := "Invalid refresh token"
        
        reqBody := `{"refresh_token": "` + token + `"}`
        req := httptest.NewRequest("POST", "/api/auth/logout", bytes.NewBufferString(reqBody))
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()

        mockDb := &mockDatabaseService{
            GetTokenFunc: func(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
                return database.RefreshToken{}, fmt.Errorf("mock error")
            },
            RevokeDelegationByIDFunc: func(ctx context.Context, id uuid.UUID) error {
                t.Fatalf("should not be called if errs")
                return nil
            },
            ExpireAllDelegationTokensFunc: func(ctx context.Context, delegationID uuid.UUID) error {
                t.Fatalf("should not be called if errs")
                return nil 
            },
        }

        mockAuth := &mockAuthService{
            HashRefreshTokenFunc: func(token string) string {
                return mockHashedToken
            },
        }

        mockApp := &handlers.AppServer{
            Db: mockDb,
            Auth: mockAuth,
            Logger: kitlog.NewNopLogger(), 
        }

        mockApp.HandlerUserLogout(rr, req)

        // --- Assertions ---
        if status := rr.Code; status != http.StatusBadRequest {
             t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusBadRequest, rr.Body.String())
            return
        }

        var actualErrorResponse handlers.ErrorResponse
        err := json.Unmarshal(rr.Body.Bytes(), &actualErrorResponse)
        if err != nil {
            t.Fatalf("could not unmarshal error response body: %v", err)
        }

        if actualErrorResponse.Error != expectedErrorMessage {
            t.Errorf("handler returned wrong error message: got %q want %q", actualErrorResponse.Error, expectedErrorMessage)
        }
    })
}