package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	kitlog "github.com/go-kit/log"
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/utils"
	"github.com/jms-guy/greed/backend/server/handlers"
	"github.com/jms-guy/greed/models"
)

func TestHandlerGetTransactionsForAccount(t *testing.T) {
    tests := []struct {
        name            string
		accountInContext any
        pathParams      map[string]string
		queryParams 	url.Values
        requestBody     string
        mockDb          *mockDatabaseService
		mockAuth 		*mockAuthService
		mockQuerier     *mockQuerier
        expectedStatus  int 
        expectedBody    any 
		sqlMockExpectations func(sqlmock.Sqlmock)
    }{
		{
            name: "should successfully get empty transaction records for user with no queries",
            accountInContext: testAccount,
			pathParams: map[string]string{"account-id": testAccountID},
			queryParams:	url.Values{},
            mockDb: &mockDatabaseService{
				GetMerchantSummaryByMonthFunc: func(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error) {
					return []database.GetMerchantSummaryByMonthRow{}, nil
				},
				GetMerchantSummaryFunc: func(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error) {
					return []database.GetMerchantSummaryRow{}, nil
				},
            },
			mockQuerier: &mockQuerier{
				ValidateQueryFunc: func(queries url.Values, rules map[string]string) (map[string]string, []utils.QueryValidationError) {
					return map[string]string{}, nil
				},
				BuildSqlQueryFunc: func(queries map[string]string, accountID string) (string, []any, error) {
					return "SELECT id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category, created_at, updated_at FROM transactions WHERE account_id = $1", []any{accountID}, nil 
				},
			},
			mockAuth: &mockAuthService{},
            expectedStatus: http.StatusOK,
            expectedBody: []models.Transaction(nil),
			sqlMockExpectations: func(mock sqlmock.Sqlmock) {
                columns := []string{"id", "account_id", "amount", "iso_currency_code", "date", "merchant_name", "payment_channel", "personal_finance_category", "created_at", "updated_at"}
                mock.ExpectQuery(`SELECT id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category, created_at, updated_at FROM transactions WHERE account_id = \$1`). 
                    WithArgs(testAccountID).
                    WillReturnRows(sqlmock.NewRows(columns)) 
            },
        },
		{
            name: "should successfully get transaction records for user with no queries",
            accountInContext: testAccount,
			pathParams: map[string]string{"account-id": testAccountID},
			queryParams:	url.Values{},
            mockDb: &mockDatabaseService{
				GetMerchantSummaryByMonthFunc: func(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error) {
					return []database.GetMerchantSummaryByMonthRow{}, nil
				},
				GetMerchantSummaryFunc: func(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error) {
					return []database.GetMerchantSummaryRow{}, nil
				},
            },
			mockQuerier: &mockQuerier{
				ValidateQueryFunc: func(queries url.Values, rules map[string]string) (map[string]string, []utils.QueryValidationError) {
					return map[string]string{}, nil
				},
				BuildSqlQueryFunc: func(queries map[string]string, accountID string) (string, []any, error) {
					return "SELECT id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category, created_at, updated_at FROM transactions WHERE account_id = $1", []any{accountID}, nil 
				},
			},
			mockAuth: &mockAuthService{},
            expectedStatus: http.StatusOK,
            expectedBody: []models.Transaction{{Id: testTxnID.String(), AccountId: testAccountID, Amount: "123.45", IsoCurrencyCode: "USD", Date: testDate.Time, MerchantName: testMerchant.String, PaymentChannel: "online", PersonalFinanceCategory: "Food & Drink"}},
			sqlMockExpectations: func(mock sqlmock.Sqlmock) {
                columns := []string{"id", "account_id", "amount", "iso_currency_code", "date", "merchant_name", "payment_channel", "personal_finance_category", "created_at", "updated_at"}
                mock.ExpectQuery(`SELECT id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category, created_at, updated_at FROM transactions WHERE account_id = \$1`). 
                    WithArgs(testAccountID).
                    WillReturnRows(sqlmock.NewRows(columns).AddRow(
							testTxnID.String(),                   
                            testAccountID,               
                            123.45,                      
                            "USD",                      
                            testDate,                    
                            testMerchant,                
                            "online",                    
                            "Food & Drink",              
                            time.Now(),                  
                            time.Now(),                  
					))
            },
        },
		{
            name: "should err with bad account in context",
			accountInContext: 1,
			pathParams: map[string]string{"account-id": testAccountID},
			queryParams:	url.Values{},
            mockDb: &mockDatabaseService{
				GetMerchantSummaryByMonthFunc: func(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error) {
					return []database.GetMerchantSummaryByMonthRow{}, nil
				},
				GetMerchantSummaryFunc: func(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error) {
					return []database.GetMerchantSummaryRow{}, nil
				},
            },
			mockQuerier: &mockQuerier{
				ValidateQueryFunc: func(queries url.Values, rules map[string]string) (map[string]string, []utils.QueryValidationError) {
					return map[string]string{}, nil
				},
				BuildSqlQueryFunc: func(queries map[string]string, accountID string) (string, []any, error) {
					return "SELECT id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category, created_at, updated_at FROM transactions WHERE account_id = $1", []any{accountID}, nil 
				},
			},
			mockAuth: &mockAuthService{},
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad account data in context",
			sqlMockExpectations: func(mock sqlmock.Sqlmock) {
                columns := []string{"id", "account_id", "amount", "iso_currency_code", "date", "merchant_name", "payment_channel", "personal_finance_category", "created_at", "updated_at"}
                mock.ExpectQuery(`SELECT id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category, created_at, updated_at FROM transactions WHERE account_id = \$1`). 
                    WithArgs(testAccountID).
                    WillReturnRows(sqlmock.NewRows(columns).AddRow(
							testTxnID.String(),                   
                            testAccountID,               
                            123.45,                      
                            "USD",                      
                            testDate,                    
                            testMerchant,                
                            "online",                    
                            "Food & Drink",              
                            time.Now(),                  
                            time.Now(),                  
					))
            },
        },
		{
            name: "should err with bad query parameter",
            accountInContext: testAccount,
			pathParams: map[string]string{"account-id": testAccountID},
			queryParams:	url.Values{"notgood": []string{"test"}},
            mockDb: &mockDatabaseService{
				GetMerchantSummaryByMonthFunc: func(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error) {
					return []database.GetMerchantSummaryByMonthRow{}, nil
				},
				GetMerchantSummaryFunc: func(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error) {
					return []database.GetMerchantSummaryRow{}, nil
				},
            },
			mockQuerier: &mockQuerier{
				ValidateQueryFunc: func(queries url.Values, rules map[string]string) (map[string]string, []utils.QueryValidationError) {
					return map[string]string{}, []utils.QueryValidationError{{Value: "fail"}}
				},
				BuildSqlQueryFunc: func(queries map[string]string, accountID string) (string, []any, error) {
					return "SELECT id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category, created_at, updated_at FROM transactions WHERE account_id = $1", []any{accountID}, nil 
				},
			},
			mockAuth: &mockAuthService{},
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad query parameter",
			sqlMockExpectations: func(mock sqlmock.Sqlmock) {
                columns := []string{"id", "account_id", "amount", "iso_currency_code", "date", "merchant_name", "payment_channel", "personal_finance_category", "created_at", "updated_at"}
                mock.ExpectQuery(`SELECT id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category, created_at, updated_at FROM transactions WHERE account_id = \$1`). 
                    WithArgs(testAccountID).
                    WillReturnRows(sqlmock.NewRows(columns).AddRow(
							testTxnID.String(),                   
                            testAccountID,               
                            123.45,                      
                            "USD",                      
                            testDate,                    
                            testMerchant,                
                            "online",                    
                            "Food & Drink",              
                            time.Now(),                  
                            time.Now(),                  
					))
            },
        },
		{
            name: "should err on building sql query",
            accountInContext: testAccount,
			pathParams: map[string]string{"account-id": testAccountID},
			queryParams:	url.Values{},
            mockDb: &mockDatabaseService{
				GetMerchantSummaryByMonthFunc: func(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error) {
					return []database.GetMerchantSummaryByMonthRow{}, nil
				},
				GetMerchantSummaryFunc: func(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error) {
					return []database.GetMerchantSummaryRow{}, nil
				},
            },
			mockQuerier: &mockQuerier{
				ValidateQueryFunc: func(queries url.Values, rules map[string]string) (map[string]string, []utils.QueryValidationError) {
					return map[string]string{}, nil
				},
				BuildSqlQueryFunc: func(queries map[string]string, accountID string) (string, []any, error) {
					return "SELECT id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category, created_at, updated_at FROM transactions WHERE account_id = $1", []any{accountID}, fmt.Errorf("mock error") 
				},
			},
			mockAuth: &mockAuthService{},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Error building database query",
			sqlMockExpectations: func(mock sqlmock.Sqlmock) {
                columns := []string{"id", "account_id", "amount", "iso_currency_code", "date", "merchant_name", "payment_channel", "personal_finance_category", "created_at", "updated_at"}
                mock.ExpectQuery(`SELECT id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category, created_at, updated_at FROM transactions WHERE account_id = \$1`). 
                    WithArgs(testAccountID).
                    WillReturnRows(sqlmock.NewRows(columns).AddRow(
							testTxnID.String(),                   
                            testAccountID,               
                            123.45,                      
                            "USD",                      
                            testDate,                    
                            testMerchant,                
                            "online",                    
                            "Food & Drink",              
                            time.Now(),                  
                            time.Now(),                  
					))
            },
        },
		{
            name: "should err on querycontext",
            accountInContext: testAccount,
			pathParams: map[string]string{"account-id": testAccountID},
			queryParams:	url.Values{},
            mockDb: &mockDatabaseService{
				GetMerchantSummaryByMonthFunc: func(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error) {
					return []database.GetMerchantSummaryByMonthRow{}, nil
				},
				GetMerchantSummaryFunc: func(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error) {
					return []database.GetMerchantSummaryRow{}, nil
				},
            },
			mockQuerier: &mockQuerier{
				ValidateQueryFunc: func(queries url.Values, rules map[string]string) (map[string]string, []utils.QueryValidationError) {
					return map[string]string{}, nil
				},
				BuildSqlQueryFunc: func(queries map[string]string, accountID string) (string, []any, error) {
					return "SELECT id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category, created_at, updated_at FROM transactions WHERE account_id = $1", []any{accountID}, nil 
				},
			},
			mockAuth: &mockAuthService{},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
			sqlMockExpectations: func(mock sqlmock.Sqlmock) {
                mock.ExpectQuery(`SELECT id, account_id, amount, iso_currency_code, date, merchant_name, payment_channel, personal_finance_category, created_at, updated_at FROM transactions WHERE account_id = \$1`). 
                    WithArgs(testAccountID).
                    WillReturnError(fmt.Errorf("mock error"))
            },
        },
		{
            name: "should successfully get summary records for user without date",
            accountInContext: testAccount,
			pathParams: map[string]string{"account-id": testAccountID},
			queryParams:	url.Values{"summary": []string{"true"}},
            mockDb: &mockDatabaseService{
				GetMerchantSummaryByMonthFunc: func(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error) {
					return []database.GetMerchantSummaryByMonthRow{}, nil
				},
				GetMerchantSummaryFunc: func(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error) {
					return []database.GetMerchantSummaryRow{{Merchant: sql.NullString{String: "test", Valid: true}, TxnCount: 5, Category: "transportation", TotalAmount: 2.00, Month: "2006-01-02"}}, nil
				},
            },
			mockQuerier: &mockQuerier{
				ValidateQueryFunc: func(queries url.Values, rules map[string]string) (map[string]string, []utils.QueryValidationError) {
					return map[string]string{"summary": "true"}, nil
				},
				BuildSqlQueryFunc: func(queries map[string]string, accountID string) (string, []any, error) {
					return  "", []any{accountID}, nil 
				},
			},
			mockAuth: &mockAuthService{},
            expectedStatus: http.StatusOK,
            expectedBody: []models.MerchantSummary{{Merchant: "test", TxnCount: 5, Category: "transportation", TotalAmount: "2.00", Month: "2006-01-02"}},
			sqlMockExpectations: nil,
        },
		{
            name: "should err on getting summary records for user without date",
            accountInContext: testAccount,
			pathParams: map[string]string{"account-id": testAccountID},
			queryParams:	url.Values{"summary": []string{"true"}},
            mockDb: &mockDatabaseService{
				GetMerchantSummaryByMonthFunc: func(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error) {
					return []database.GetMerchantSummaryByMonthRow{}, nil
				},
				GetMerchantSummaryFunc: func(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error) {
					return []database.GetMerchantSummaryRow{}, fmt.Errorf("mock error")
				},
            },
			mockQuerier: &mockQuerier{
				ValidateQueryFunc: func(queries url.Values, rules map[string]string) (map[string]string, []utils.QueryValidationError) {
					return map[string]string{"summary": "true"}, nil
				},
				BuildSqlQueryFunc: func(queries map[string]string, accountID string) (string, []any, error) {
					return  "", []any{accountID}, nil 
				},
			},
			mockAuth: &mockAuthService{},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
			sqlMockExpectations: nil,
        },
		{
            name: "should successfully get summary records for user with date",
            accountInContext: testAccount,
			pathParams: map[string]string{"account-id": testAccountID},
			queryParams:	url.Values{"summary": []string{"true"}, "date": []string{"2006-01-02"}},
            mockDb: &mockDatabaseService{
				GetMerchantSummaryByMonthFunc: func(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error) {
					return []database.GetMerchantSummaryByMonthRow{{Merchant: sql.NullString{String: "test", Valid: true}, TxnCount: 5, Category: "transportation", TotalAmount: 2.00, Month: "2006-01-02"}}, nil
				},
				GetMerchantSummaryFunc: func(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error) {
					return []database.GetMerchantSummaryRow{}, nil
				},
            },
			mockQuerier: &mockQuerier{
				ValidateQueryFunc: func(queries url.Values, rules map[string]string) (map[string]string, []utils.QueryValidationError) {
					return map[string]string{"summary": "true", "date": "2006-01-02"}, nil
				},
				BuildSqlQueryFunc: func(queries map[string]string, accountID string) (string, []any, error) {
					return  "", []any{accountID}, nil 
				},
			},
			mockAuth: &mockAuthService{},
            expectedStatus: http.StatusOK,
            expectedBody: []models.MerchantSummary{{Merchant: "test", TxnCount: 5, Category: "transportation", TotalAmount: "2.00", Month: "2006-01-02"}},
			sqlMockExpectations: nil,
        },
		{
            name: "should err on getting summary records for user with date",
            accountInContext: testAccount,
			pathParams: map[string]string{"account-id": testAccountID},
			queryParams:	url.Values{"summary": []string{"true"}, "date": []string{"2006-01-02"}},
            mockDb: &mockDatabaseService{
				GetMerchantSummaryByMonthFunc: func(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error) {
					return []database.GetMerchantSummaryByMonthRow{}, fmt.Errorf("mock error")
				},
				GetMerchantSummaryFunc: func(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error) {
					return []database.GetMerchantSummaryRow{}, nil
				},
            },
			mockQuerier: &mockQuerier{
				ValidateQueryFunc: func(queries url.Values, rules map[string]string) (map[string]string, []utils.QueryValidationError) {
					return map[string]string{"summary": "true", "date": "2006-01-02"}, nil
				},
				BuildSqlQueryFunc: func(queries map[string]string, accountID string) (string, []any, error) {
					return  "", []any{accountID}, nil 
				},
			},
			mockAuth: &mockAuthService{},
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
			sqlMockExpectations: nil,
        },
	}

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
                t.Fatalf("an error '%s' occurred when opening a mock database connection", err)
            }
            defer mockDB.Close()

			if tt.sqlMockExpectations != nil {
                tt.sqlMockExpectations(mock)
            }

			reqURL := fmt.Sprintf("/api/accounts/%s/transactions", tt.pathParams["account-id"])
            
            if len(tt.queryParams) > 0 {
                reqURL += "?" + tt.queryParams.Encode()
            }

            req := httptest.NewRequest("GET", reqURL, bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            
            rctx := chi.NewRouteContext()
            for key, value := range tt.pathParams {
                rctx.URLParams.Add(key, value)
            }

            ctx := context.WithValue(req.Context(),handlers.GetAccountKey(), tt.accountInContext)

            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
				Database: mockDB,
                Logger: kitlog.NewNopLogger(),
				Auth: tt.mockAuth,
				Config: &config.Config{},
				Querier: tt.mockQuerier,
            }

            mockApp.HandlerGetTransactionsForAccount(rr, req)

            // --- Assertions ---
            if status := rr.Code; status != tt.expectedStatus {
                t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, tt.expectedStatus, rr.Body.String())
            }

            switch expected := tt.expectedBody.(type) {
            case []models.Transaction:
                var actual []models.Transaction
                if err := json.Unmarshal(rr.Body.Bytes(), &actual); err != nil {
                    t.Fatalf("Failed to unmarshal response body to []models.Transaction: %v, Body: %s", err, rr.Body.String())
                }
				for i := range actual {
                    actual[i].Date = actual[i].Date.Truncate(0) 
                }
                for i := range expected {
                    expected[i].Date = expected[i].Date.Truncate(0) 
                }

                if !reflect.DeepEqual(actual, expected) {
                    t.Errorf("handler returned unexpected transactions: got %+v want %+v", actual, expected)
                }
            case []models.MerchantSummary:
                var actual []models.MerchantSummary
                if err := json.Unmarshal(rr.Body.Bytes(), &actual); err != nil {
                    t.Fatalf("Failed to unmarshal response body to []models.MerchantSummary: %v, Body: %s", err, rr.Body.String())
                }
                if !reflect.DeepEqual(actual, expected) {
                    t.Errorf("handler returned unexpected merchant summaries: got %+v want %+v", actual, expected)
                }
            case string:
                if !strings.Contains(rr.Body.String(), expected) {
                    t.Errorf("handler returned unexpected body: got %s want body to contain %s", rr.Body.String(), expected)
                }
            default:
                t.Fatalf("Unsupported expectedBody type: %T", expected)
            }
    	})
	}
}

func TestHandlerGetMonetaryDataForMonth(t *testing.T) {
    tests := []struct {
        name            string
		accountInContext any
		pathParams 		map[string]string
        requestBody     string
        mockDb          *mockDatabaseService
		mockAuth 		*mockAuthService
        expectedStatus  int 
        expectedBody    any  
    }{
		{
            name: "should successfully return items for user with credit type",
			accountInContext: database.Account{ID: testAccountID, Type: "credit"},
			pathParams: map[string]string{"year": "2006", "month": "10", "account-id": testAccountID},
            mockDb: &mockDatabaseService{
                GetMonetaryDataForMonthFunc: func(ctx context.Context, arg database.GetMonetaryDataForMonthParams) (database.GetMonetaryDataForMonthRow, error) {
					return database.GetMonetaryDataForMonthRow{Income: "100", Expenses: "0", NetIncome: "100"}, nil
				},
            },
            expectedStatus: http.StatusOK,
            expectedBody: models.MonetaryData{Income: "100", Expenses: "0", NetIncome: "100", Date: "2006-10"},
        },
		{
            name: "should successfully return items for user with depository type",
			accountInContext: database.Account{ID: testAccountID, Type: "depository"},
			pathParams: map[string]string{"year": "2006", "month": "10", "account-id": testAccountID},
            mockDb: &mockDatabaseService{
                GetMonetaryDataForMonthFunc: func(ctx context.Context, arg database.GetMonetaryDataForMonthParams) (database.GetMonetaryDataForMonthRow, error) {
					return database.GetMonetaryDataForMonthRow{Income: "100", Expenses: "0", NetIncome: "100"}, nil
				},
            },
            expectedStatus: http.StatusOK,
            expectedBody: models.MonetaryData{Income: "100", Expenses: "0", NetIncome: "100", Date: "2006-10"},
        },
		{
            name: "should err with bad account in context",
			accountInContext: 1,
			pathParams: map[string]string{"year": "2006", "month": "10", "account-id": testAccountID},
            mockDb: &mockDatabaseService{
                GetMonetaryDataForMonthFunc: func(ctx context.Context, arg database.GetMonetaryDataForMonthParams) (database.GetMonetaryDataForMonthRow, error) {
					return database.GetMonetaryDataForMonthRow{Income: "100", Expenses: "0", NetIncome: "100"}, nil
				},
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad account in context",
        },
		{
            name: "should err on getting database data",
			accountInContext: database.Account{ID: testAccountID, Type: "depository"},
			pathParams: map[string]string{"year": "2006", "month": "10", "account-id": testAccountID},
            mockDb: &mockDatabaseService{
                GetMonetaryDataForMonthFunc: func(ctx context.Context, arg database.GetMonetaryDataForMonthParams) (database.GetMonetaryDataForMonthRow, error) {
					return database.GetMonetaryDataForMonthRow{}, fmt.Errorf("mock error")
				},
            },
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
		{
            name: "should err with invalid month format",
			accountInContext: database.Account{ID: testAccountID, Type: "depository"},
			pathParams: map[string]string{"year": "2006", "month": "15", "account-id": testAccountID},
            mockDb: &mockDatabaseService{
                GetMonetaryDataForMonthFunc: func(ctx context.Context, arg database.GetMonetaryDataForMonthParams) (database.GetMonetaryDataForMonthRow, error) {
					return database.GetMonetaryDataForMonthRow{Income: "100", Expenses: "0", NetIncome: "100"}, nil
				},
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Invalid month format",
        },
		{
            name: "should err with invalid year format",
			accountInContext: database.Account{ID: testAccountID, Type: "depository"},
			pathParams: map[string]string{"year": "yes", "month": "10", "account-id": testAccountID},
            mockDb: &mockDatabaseService{
                GetMonetaryDataForMonthFunc: func(ctx context.Context, arg database.GetMonetaryDataForMonthParams) (database.GetMonetaryDataForMonthRow, error) {
					return database.GetMonetaryDataForMonthRow{Income: "100", Expenses: "0", NetIncome: "100"}, nil
				},
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Invalid year format",
        },
		{
            name: "should err on account with wrong type",
			accountInContext: database.Account{ID: testAccountID, Type: "loan"},
			pathParams: map[string]string{"year": "2006", "month": "10", "account-id": testAccountID},
            mockDb: &mockDatabaseService{
                GetMonetaryDataForMonthFunc: func(ctx context.Context, arg database.GetMonetaryDataForMonthParams) (database.GetMonetaryDataForMonthRow, error) {
					return database.GetMonetaryDataForMonthRow{Income: "100", Expenses: "0", NetIncome: "100"}, nil
				},
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Endpoint only works with debit or credit",
        },
	}

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
			reqURL := fmt.Sprintf("/api/accounts/%s/transactions/monetary/%s-%s", tt.pathParams["account-id"], tt.pathParams["year"], tt.pathParams["month"])

            req := httptest.NewRequest("GET", reqURL, bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            
             rctx := chi.NewRouteContext()
            rctx.URLParams.Add("account-id", tt.pathParams["account-id"])
            rctx.URLParams.Add("year", tt.pathParams["year"])
            rctx.URLParams.Add("month", tt.pathParams["month"])

            ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
            ctx = context.WithValue(ctx, handlers.GetAccountKey(), tt.accountInContext)
            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
                Logger: kitlog.NewNopLogger(),
            }

            mockApp.HandlerGetMonetaryDataForMonth(rr, req)

            // --- Assertions ---
            if status := rr.Code; status != tt.expectedStatus {
                t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, tt.expectedStatus, rr.Body.String())
            }

			switch expected := tt.expectedBody.(type) {
            case models.MonetaryData:
                var actual models.MonetaryData
                if err := json.Unmarshal(rr.Body.Bytes(), &actual); err != nil {
                    t.Fatalf("Failed to unmarshal response body to []models.MerchantSummary: %v, Body: %s", err, rr.Body.String())
                }
                if !reflect.DeepEqual(actual, expected) {
                    t.Errorf("handler returned unexpected merchant summaries: got %+v want %+v", actual, expected)
                }
            case string:
                if !strings.Contains(rr.Body.String(), expected) {
                    t.Errorf("handler returned unexpected body: got %s want body to contain %s", rr.Body.String(), expected)
                }
            default:
                t.Fatalf("Unsupported expectedBody type: %T", expected)
            }
        })
    }
}

func TestHandlerGetMonetaryData(t *testing.T) {
    tests := []struct {
        name            string
		accountInContext any
		pathParams 		map[string]string
        requestBody     string
        mockDb          *mockDatabaseService
		mockAuth 		*mockAuthService
        expectedStatus  int 
        expectedBody    any  
    }{
		{
            name: "should successfully return items for user with credit type",
			accountInContext: database.Account{ID: testAccountID, Type: "credit"},
			pathParams: map[string]string{"account-id": testAccountID},
            mockDb: &mockDatabaseService{
                GetMonetaryDataForAllMonthsFunc: func(ctx context.Context, accountID string) ([]database.GetMonetaryDataForAllMonthsRow, error) {
					return []database.GetMonetaryDataForAllMonthsRow{{Income: "100", Expenses: "0", NetIncome: "100"}}, nil
				},
            },
            expectedStatus: http.StatusOK,
            expectedBody: []models.MonetaryData{{Income: "100", Expenses: "0", NetIncome: "100", Date: "0-0"}},
        },
		{
            name: "should successfully return items for user with depository type",
			accountInContext: database.Account{ID: testAccountID, Type: "depository"},
			pathParams: map[string]string{"account-id": testAccountID},
            mockDb: &mockDatabaseService{
                GetMonetaryDataForAllMonthsFunc: func(ctx context.Context, accountID string) ([]database.GetMonetaryDataForAllMonthsRow, error) {
					return []database.GetMonetaryDataForAllMonthsRow{{Income: "100", Expenses: "0", NetIncome: "100"}}, nil
				},
            },
            expectedStatus: http.StatusOK,
            expectedBody: []models.MonetaryData{{Income: "100", Expenses: "0", NetIncome: "100", Date: "0-0"}},
        },
		{
            name: "should err for account with bad type",
			accountInContext: database.Account{ID: testAccountID, Type: "loan"},
			pathParams: map[string]string{"account-id": testAccountID},
            mockDb: &mockDatabaseService{
                GetMonetaryDataForAllMonthsFunc: func(ctx context.Context, accountID string) ([]database.GetMonetaryDataForAllMonthsRow, error) {
					return []database.GetMonetaryDataForAllMonthsRow{{Income: "100", Expenses: "0", NetIncome: "100"}}, nil
				},
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Endpoint only works",
        },
		{
            name: "should err with bad account in context",
			accountInContext: 1,
			pathParams: map[string]string{"account-id": testAccountID},
            mockDb: &mockDatabaseService{
                GetMonetaryDataForAllMonthsFunc: func(ctx context.Context, accountID string) ([]database.GetMonetaryDataForAllMonthsRow, error) {
					return []database.GetMonetaryDataForAllMonthsRow{{Income: "100", Expenses: "0", NetIncome: "100"}}, nil
				},
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad account in context",
        },
		{
            name: "should err getting data from database",
			accountInContext: database.Account{ID: testAccountID, Type: "credit"},
			pathParams: map[string]string{"account-id": testAccountID},
            mockDb: &mockDatabaseService{
                GetMonetaryDataForAllMonthsFunc: func(ctx context.Context, accountID string) ([]database.GetMonetaryDataForAllMonthsRow, error) {
					return []database.GetMonetaryDataForAllMonthsRow{}, fmt.Errorf("mock error")
				},
            },
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
	}

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
			reqURL := fmt.Sprintf("/api/accounts/%s/transactions/monetary", tt.pathParams["account-id"])

            req := httptest.NewRequest("GET", reqURL, bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            
            rctx := chi.NewRouteContext()
            rctx.URLParams.Add("account-id", tt.pathParams["account-id"])

            ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
            ctx = context.WithValue(ctx, handlers.GetAccountKey(), tt.accountInContext)
            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
                Logger: kitlog.NewNopLogger(),
            }

            mockApp.HandlerGetMonetaryData(rr, req)

            // --- Assertions ---
            if status := rr.Code; status != tt.expectedStatus {
                t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, tt.expectedStatus, rr.Body.String())
            }

			switch expected := tt.expectedBody.(type) {
            case []models.MonetaryData:
                var actual []models.MonetaryData
                if err := json.Unmarshal(rr.Body.Bytes(), &actual); err != nil {
                    t.Fatalf("Failed to unmarshal response body to []models.MerchantSummary: %v, Body: %s", err, rr.Body.String())
                }
                if !reflect.DeepEqual(actual, expected) {
                    t.Errorf("handler returned unexpected merchant summaries: got %+v want %+v", actual, expected)
                }
            case string:
                if !strings.Contains(rr.Body.String(), expected) {
                    t.Errorf("handler returned unexpected body: got %s want body to contain %s", rr.Body.String(), expected)
                }
            default:
                t.Fatalf("Unsupported expectedBody type: %T", expected)
            }
        })
    }
}

func TestHandlerDeleteTransactionsForAccount(t *testing.T) {
    tests := []struct {
        name            string
		accountInContext any
		pathParams 		map[string]string
        requestBody     string
        mockDb          *mockDatabaseService
		mockAuth 		*mockAuthService
        expectedStatus  int 
        expectedBody    string  
    }{
		{
            name: "should successfully delete transactions for account",
			accountInContext: database.Account{ID: testAccountID},
			pathParams: map[string]string{"account-id": testAccountID},
            mockDb: &mockDatabaseService{
                DeleteTransactionsForAccountFunc: func(ctx context.Context, accountID string) error {
					return nil
				},
            },
            expectedStatus: http.StatusOK,
            expectedBody: "Transactions deleted successfully",
        },
		{
            name: "should err with bad account in context",
			accountInContext: 1,
			pathParams: map[string]string{"account-id": testAccountID},
            mockDb: &mockDatabaseService{
                DeleteTransactionsForAccountFunc: func(ctx context.Context, accountID string) error {
					return nil
				},
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: "Bad account in context",
        },
		{
            name: "should err deleting transactions",
			accountInContext: database.Account{ID: testAccountID},
			pathParams: map[string]string{"account-id": testAccountID},
            mockDb: &mockDatabaseService{
                DeleteTransactionsForAccountFunc: func(ctx context.Context, accountID string) error {
					return fmt.Errorf("mock error")
				},
            },
            expectedStatus: http.StatusInternalServerError,
            expectedBody: "Database error",
        },
	}

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
			reqURL := fmt.Sprintf("/api/accounts/%s/transactions/monetary", tt.pathParams["account-id"])

            req := httptest.NewRequest("GET", reqURL, bytes.NewBufferString(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            
            rctx := chi.NewRouteContext()
            rctx.URLParams.Add("account-id", tt.pathParams["account-id"])

            ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
            ctx = context.WithValue(ctx, handlers.GetAccountKey(), tt.accountInContext)
            req = req.WithContext(ctx)

            rr := httptest.NewRecorder()

            mockApp := &handlers.AppServer{
                Db: tt.mockDb,
                Logger: kitlog.NewNopLogger(),
            }

            mockApp.HandlerDeleteTransactionsForAccount(rr, req)

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