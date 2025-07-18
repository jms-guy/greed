package handlers_test

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/api/sgrid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/jms-guy/greed/backend/internal/database"
)

//Testing global variables
var (
    testUserID = uuid.MustParse("a1b2c3d4-e5f6-7890-1234-567890abcdef")
	testName = "Test"
	testEmail = "test@email.com"
	testItemName = "ItemName"
	testAccountID = "54321"
	testAccountName = "AccountName"
	testAccessToken = "testAccessToken"
)

//Test database service
type mockDatabaseService struct {
    GetUserByNameFunc           func(ctx context.Context, name string) (database.User, error)
    CreateUserFunc              func(ctx context.Context, params database.CreateUserParams) (database.User, error)
	DeleteUserFunc              func(ctx context.Context, id uuid.UUID) error
	GetAllUsersFunc             func(ctx context.Context) ([]string, error)
	GetUserFunc                 func(ctx context.Context, id uuid.UUID) (database.User, error)
	GetUserByEmailFunc          func(ctx context.Context, email string) (database.User, error)
	ResetUsersFunc              func(ctx context.Context) error 
	UpdateFreeCallsFunc         func(ctx context.Context, id uuid.UUID) error
	UpdateMemberFunc            func(ctx context.Context, id uuid.UUID) error
	UpdatePasswordFunc          func(ctx context.Context, arg database.UpdatePasswordParams) error 
	VerifyUserFunc              func(ctx context.Context, id uuid.UUID) error
	GetAllAccountsForUserFunc   func(ctx context.Context, userID uuid.UUID) ([]database.Account, error)
	CreateAccountFunc           func(ctx context.Context, arg database.CreateAccountParams) (database.Account, error)
	DeleteAccountFunc           func(ctx context.Context, arg database.DeleteAccountParams) error
	GetAccountFunc              func(ctx context.Context, name string) (database.Account, error)
	GetAccountByIdFunc          func(ctx context.Context, arg database.GetAccountByIdParams) (database.Account, error)
	GetAccountsForItemFunc      func(ctx context.Context, itemID string) ([]database.Account, error)
	ResetAccountsFunc           func(ctx context.Context) error
	UpdateBalancesFunc          func(ctx context.Context, arg database.UpdateBalancesParams) (database.Account, error)
	GetMerchantSummaryFunc              func(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error)
	GetMerchantSummaryByMonthFunc       func(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error)
	GetMonetaryDataForAllMonthsFunc     func(ctx context.Context, accountID string) ([]database.GetMonetaryDataForAllMonthsRow, error)
	GetMonetaryDataForMonthFunc         func(ctx context.Context, arg database.GetMonetaryDataForMonthParams) (database.GetMonetaryDataForMonthRow, error)
	ValidateCurrencyFunc                func(ctx context.Context, code string) (bool, error)
	CreateDelegationFunc                func(ctx context.Context, arg database.CreateDelegationParams) (database.Delegation, error)
	GetDelegationFunc                   func(ctx context.Context, id uuid.UUID) (database.Delegation, error)
	RevokeDelegationByIDFunc            func(ctx context.Context, id uuid.UUID) error
	RevokeDelegationByUserFunc          func(ctx context.Context, userID uuid.UUID) error
	UpdateLastUsedFunc                  func(ctx context.Context, id uuid.UUID) error
	CreateItemFunc                      func(ctx context.Context, arg database.CreateItemParams) (database.PlaidItem, error)
	DeleteItemFunc                      func(ctx context.Context, arg database.DeleteItemParams) error
	GetAccessTokenFunc                  func(ctx context.Context, id string) (database.PlaidItem, error)
	GetCursorFunc                       func(ctx context.Context, arg database.GetCursorParams) (sql.NullString, error)
	GetItemByIDFunc                     func(ctx context.Context, id string) (database.PlaidItem, error)
	GetItemByNameFunc                   func(ctx context.Context, nickname sql.NullString) (database.PlaidItem, error)
	GetItemsByUserFunc                  func(ctx context.Context, userID uuid.UUID) ([]database.PlaidItem, error)
	GetLatestCursorOrNilFunc            func(ctx context.Context, id string) (sql.NullString, error)
	ResetItemsFunc                      func(ctx context.Context) error
	UpdateCursorFunc                    func(ctx context.Context, arg database.UpdateCursorParams) error
	UpdateNicknameFunc                  func(ctx context.Context, arg database.UpdateNicknameParams) error
	CreateTokenFunc                     func(ctx context.Context, arg database.CreateTokenParams) (database.RefreshToken, error)
	ExpireAllDelegationTokensFunc       func(ctx context.Context, delegationID uuid.UUID) error
	ExpireTokenFunc                     func(ctx context.Context, hashedToken string) error
	GetTokenFunc                        func(ctx context.Context, hashedToken string) (database.RefreshToken, error)
	ClearTransactionsTableFunc          func(ctx context.Context) error
	CreateTransactionFunc               func(ctx context.Context, arg database.CreateTransactionParams) (database.Transaction, error)
	DeleteTransactionFunc               func(ctx context.Context, arg database.DeleteTransactionParams) error
	DeleteTransactionsForAccountFunc    func(ctx context.Context, accountID string) error
	GetTransactionsFunc                 func(ctx context.Context, accountID string) ([]database.Transaction, error)
	GetTransactionsForUserFunc          func(ctx context.Context, userID uuid.UUID) ([]database.Transaction, error)
	CreateVerificationRecordFunc        func(ctx context.Context, arg database.CreateVerificationRecordParams) (database.VerificationRecord, error)
	DeleteVerificationRecordFunc        func(ctx context.Context, verificationCode string) error
	DeleteVerificationRecordByUserFunc  func(ctx context.Context, userID uuid.UUID) error
	GetVerificationRecordFunc           func(ctx context.Context, verificationCode string) (database.VerificationRecord, error)
	GetVerificationRecordByUserFunc     func(ctx context.Context, userID uuid.UUID) (database.VerificationRecord, error)
    WithTxFunc                          func(tx *sql.Tx) *database.Queries
}

//Test Auth service
type mockAuthService struct {
    EmailValidationFunc         func(email string) bool
    HashPasswordFunc            func(password string) (string, error)
    ValidatePasswordHashFunc    func(hash, password string) error
    GenerateCodeFunc            func() string
    GetBearerTokenFunc          func(headers http.Header) (string, error)
	MakeJWTFunc                 func(cfg *config.Config, userID uuid.UUID) (string, error)
	ValidateJWTFunc             func(cfg *config.Config, tokenString string) (uuid.UUID, error)
    MakeRefreshTokenFunc        func(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error)
	HashRefreshTokenFunc        func(token string) string
}

//Test Mail service
type mockMailService struct {
	NewMailFunc 				func(from string, to string, subject string, body string, data *sgrid.MailData) *sgrid.Mail
	SendMailFunc 				func(mailreq *sgrid.Mail) error
}

type mockPlaidService struct