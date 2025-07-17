package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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