package handlers

import (
	"context"
	"database/sql"
	"net/http"

	kitlog "github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/api/sgrid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/limiter"
	"github.com/plaid/plaid-go/v36/plaid"
)

// Holds state of important server structs
type AppServer struct {
	Db       	GreedDatabase      		//SQL database queries
	Auth 		AuthService				//Auth service interface
	Database 	*sql.DB                 //SQL database
	Config   	*config.Config          //Environment variables configured from .env file
	Logger   	kitlog.Logger           //Logging interface
	SgMail   	MailService    			//SendGrid mail service
	Limiter  	*limiter.IPRateLimiter  //Rate limiter
	PClient  	*plaid.APIClient        //Client for Plaid integration
}

//Interfaces created as placeholders in the server struct, so that mock services may be created in testing that can replace actual services

//Database interface 
type GreedDatabase interface {
    GetUserByName(ctx context.Context, name string) (database.User, error)
    CreateUser(ctx context.Context, params database.CreateUserParams) (database.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetAllUsers(ctx context.Context) ([]string, error)
	GetUser(ctx context.Context, id uuid.UUID) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
	ResetUsers(ctx context.Context) error 
	UpdateFreeCalls(ctx context.Context, id uuid.UUID) error
	UpdateMember(ctx context.Context, id uuid.UUID) error
	UpdatePassword(ctx context.Context, arg database.UpdatePasswordParams) error 
	VerifyUser(ctx context.Context, id uuid.UUID) error
	GetAllAccountsForUser(ctx context.Context, userID uuid.UUID) ([]database.Account, error)
	CreateAccount(ctx context.Context, arg database.CreateAccountParams) (database.Account, error)
	DeleteAccount(ctx context.Context, arg database.DeleteAccountParams) error
	GetAccount(ctx context.Context, name string) (database.Account, error)
	GetAccountById(ctx context.Context, arg database.GetAccountByIdParams) (database.Account, error)
	GetAccountsForItem(ctx context.Context, itemID string) ([]database.Account, error)
	ResetAccounts(ctx context.Context) error
	UpdateBalances(ctx context.Context, arg database.UpdateBalancesParams) (database.Account, error)
	GetMerchantSummary(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error)
	GetMerchantSummaryByMonth(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error)
	GetMonetaryDataForAllMonths(ctx context.Context, accountID string) ([]database.GetMonetaryDataForAllMonthsRow, error)
	GetMonetaryDataForMonth(ctx context.Context, arg database.GetMonetaryDataForMonthParams) (database.GetMonetaryDataForMonthRow, error)
	ValidateCurrency(ctx context.Context, code string) (bool, error)
	CreateDelegation(ctx context.Context, arg database.CreateDelegationParams) (database.Delegation, error)
	GetDelegation(ctx context.Context, id uuid.UUID) (database.Delegation, error)
	RevokeDelegationByID(ctx context.Context, id uuid.UUID) error
	RevokeDelegationByUser(ctx context.Context, userID uuid.UUID) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	CreateItem(ctx context.Context, arg database.CreateItemParams) (database.PlaidItem, error)
	DeleteItem(ctx context.Context, arg database.DeleteItemParams) error
	GetAccessToken(ctx context.Context, id string) (database.PlaidItem, error)
	GetCursor(ctx context.Context, arg database.GetCursorParams) (sql.NullString, error)
	GetItemByID(ctx context.Context, id string) (database.PlaidItem, error)
	GetItemByName(ctx context.Context, nickname sql.NullString) (database.PlaidItem, error)
	GetItemsByUser(ctx context.Context, userID uuid.UUID) ([]database.PlaidItem, error)
	GetLatestCursorOrNil(ctx context.Context, id string) (sql.NullString, error)
	ResetItems(ctx context.Context) error
	UpdateCursor(ctx context.Context, arg database.UpdateCursorParams) error
	UpdateNickname(ctx context.Context, arg database.UpdateNicknameParams) error
	CreateToken(ctx context.Context, arg database.CreateTokenParams) (database.RefreshToken, error)
	ExpireAllDelegationTokens(ctx context.Context, delegationID uuid.UUID) error
	ExpireToken(ctx context.Context, hashedToken string) error
	GetToken(ctx context.Context, hashedToken string) (database.RefreshToken, error)
	ClearTransactionsTable(ctx context.Context) error
	CreateTransaction(ctx context.Context, arg database.CreateTransactionParams) (database.Transaction, error)
	DeleteTransaction(ctx context.Context, arg database.DeleteTransactionParams) error
	DeleteTransactionsForAccount(ctx context.Context, accountID string) error
	GetTransactions(ctx context.Context, accountID string) ([]database.Transaction, error)
	GetTransactionsForUser(ctx context.Context, userID uuid.UUID) ([]database.Transaction, error)
	CreateVerificationRecord(ctx context.Context, arg database.CreateVerificationRecordParams) (database.VerificationRecord, error)
	DeleteVerificationRecord(ctx context.Context, verificationCode string) error
	DeleteVerificationRecordByUser(ctx context.Context, userID uuid.UUID) error
	GetVerificationRecord(ctx context.Context, verificationCode string) (database.VerificationRecord, error)
	GetVerificationRecordByUser(ctx context.Context, userID uuid.UUID) (database.VerificationRecord, error)
	WithTx(tx *sql.Tx) *database.Queries
}

//Auth interface
type AuthService interface {
	EmailValidation(email string) bool
	ValidatePasswordHash(hash, password string) error
	HashPassword(password string) (string, error)
	GenerateCode() string
	GetBearerToken(headers http.Header) (string, error)
	MakeJWT(cfg *config.Config, userID uuid.UUID) (string, error)
	ValidateJWT(cfg *config.Config, tokenString string) (uuid.UUID, error)
	MakeRefreshToken(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error)
	HashRefreshToken(token string) string
}

//Sendgrid interface
type MailService interface {
	NewMail(from string, to string, subject string, body string, data *sgrid.MailData) *sgrid.Mail
	SendMail(mailreq *sgrid.Mail) error
}