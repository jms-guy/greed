package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"github.com/jms-guy/greed/backend/server/handlers"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
)

func TestHandlerCreateUser(t *testing.T) {
	reqBody := `{"name": "testuser", "email": "test@example.com", "password": "password123"}`
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

    mockDb := &mockDatabaseService{
        GetUserByNameFunc: func(ctx context.Context, name string) (database.User, error) {
            return database.User{}, sql.ErrNoRows // Simulate user not found 
        },
        CreateUserFunc: func(ctx context.Context, params database.CreateUserParams) (database.User, error) {
            // Simulate successful creation
            return database.User{
                ID: uuid.New(), Name: params.Name, HashedPassword: params.HashedPassword,
                Email: params.Email, CreatedAt: time.Now(), UpdatedAt: time.Now(),
            }, nil
        },
    }

    mockApp := &handlers.AppServer{
        Db: mockDb, // THIS WILL NOW WORK because mockDb implements handlers.GreedDatabase!
        // You'll also need to provide mocks or actual instances for any other fields in AppServer
        // that HandlerCreateUser might use (e.g., Logger, SgMail, Limiter).
        // For example, if respondWithError/respondWithJSON are methods on AppServer that use Logger:
        // Logger: kitlog.NewNopLogger(), // Use a no-op logger for tests
    }

    mockApp.HandlerCreateUser(rr, req)

	if status := rr.Code; status != http.StatusCreated {
    t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}
}