package handlers_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/api/sgrid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/utils"
	"github.com/jms-guy/greed/models"
	"github.com/plaid/plaid-go/v36/plaid"
)

func (m *mockDatabaseService) WithTx(tx *sql.Tx) *database.Queries {
	if m.WithTxFunc != nil {
		return m.WithTxFunc(tx)
	}
	return &database.Queries{}
}

func (m *mockDatabaseService) GetUserByName(ctx context.Context, name string) (database.User, error) {
	if m.GetUserByNameFunc != nil {
		return m.GetUserByNameFunc(ctx, name)
	}
	return database.User{}, nil
}

func (m *mockDatabaseService) CreateUser(ctx context.Context, params database.CreateUserParams) (database.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, params)
	}
	return database.User{}, nil
}

func (m *mockDatabaseService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}

func (m *mockDatabaseService) GetAllUsers(ctx context.Context) ([]string, error) {
	if m.GetAllUsersFunc != nil {
		return m.GetAllUsersFunc(ctx)
	}
	return nil, nil
}

func (m *mockDatabaseService) GetUser(ctx context.Context, id uuid.UUID) (database.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, id)
	}
	return database.User{}, nil
}

func (m *mockDatabaseService) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return database.User{}, nil
}

func (m *mockDatabaseService) ResetUsers(ctx context.Context) error {
	if m.ResetUsersFunc != nil {
		return m.ResetUsersFunc(ctx)
	}
	return nil
}

func (m *mockDatabaseService) UpdateFreeCalls(ctx context.Context, id uuid.UUID) error {
	if m.UpdateFreeCallsFunc != nil {
		return m.UpdateFreeCallsFunc(ctx, id)
	}
	return nil
}

func (m *mockDatabaseService) UpdateMember(ctx context.Context, id uuid.UUID) error {
	if m.UpdateMemberFunc != nil {
		return m.UpdateMemberFunc(ctx, id)
	}
	return nil
}

func (m *mockDatabaseService) UpdatePassword(ctx context.Context, arg database.UpdatePasswordParams) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(ctx, arg)
	}
	return nil
}

func (m *mockDatabaseService) VerifyUser(ctx context.Context, id uuid.UUID) error {
	if m.VerifyUserFunc != nil {
		return m.VerifyUserFunc(ctx, id)
	}
	return nil
}

func (m *mockDatabaseService) GetAllAccountsForUser(ctx context.Context, userID uuid.UUID) ([]database.Account, error) {
	if m.GetAllAccountsForUserFunc != nil {
		return m.GetAllAccountsForUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockDatabaseService) CreateAccount(ctx context.Context, arg database.CreateAccountParams) (database.Account, error) {
	if m.CreateAccountFunc != nil {
		return m.CreateAccountFunc(ctx, arg)
	}
	return database.Account{}, nil
}

func (m *mockDatabaseService) DeleteAccount(ctx context.Context, arg database.DeleteAccountParams) error {
	if m.DeleteAccountFunc != nil {
		return m.DeleteAccountFunc(ctx, arg)
	}
	return nil
}

func (m *mockDatabaseService) GetAccount(ctx context.Context, name string) (database.Account, error) {
	if m.GetAccountFunc != nil {
		return m.GetAccountFunc(ctx, name)
	}
	return database.Account{}, nil
}

func (m *mockDatabaseService) GetAccountById(ctx context.Context, arg database.GetAccountByIdParams) (database.Account, error) {
	if m.GetAccountByIdFunc != nil {
		return m.GetAccountByIdFunc(ctx, arg)
	}
	return database.Account{}, nil
}

func (m *mockDatabaseService) GetAccountsForItem(ctx context.Context, itemID string) ([]database.Account, error) {
	if m.GetAccountsForItemFunc != nil {
		return m.GetAccountsForItemFunc(ctx, itemID)
	}
	return nil, nil
}

func (m *mockDatabaseService) ResetAccounts(ctx context.Context) error {
	if m.ResetAccountsFunc != nil {
		return m.ResetAccountsFunc(ctx)
	}
	return nil
}

func (m *mockDatabaseService) UpdateBalances(ctx context.Context, arg database.UpdateBalancesParams) (database.Account, error) {
	if m.UpdateBalancesFunc != nil {
		return m.UpdateBalancesFunc(ctx, arg)
	}
	return database.Account{}, nil
}

func (m *mockDatabaseService) GetMerchantSummary(ctx context.Context, accountID string) ([]database.GetMerchantSummaryRow, error) {
	if m.GetMerchantSummaryFunc != nil {
		return m.GetMerchantSummaryFunc(ctx, accountID)
	}
	return nil, nil
}

func (m *mockDatabaseService) GetMerchantSummaryByMonth(ctx context.Context, arg database.GetMerchantSummaryByMonthParams) ([]database.GetMerchantSummaryByMonthRow, error) {
	if m.GetMerchantSummaryByMonthFunc != nil {
		return m.GetMerchantSummaryByMonthFunc(ctx, arg)
	}
	return nil, nil
}

func (m *mockDatabaseService) GetMonetaryDataForAllMonths(ctx context.Context, accountID string) ([]database.GetMonetaryDataForAllMonthsRow, error) {
	if m.GetMonetaryDataForAllMonthsFunc != nil {
		return m.GetMonetaryDataForAllMonthsFunc(ctx, accountID)
	}
	return nil, nil
}

func (m *mockDatabaseService) GetMonetaryDataForMonth(ctx context.Context, arg database.GetMonetaryDataForMonthParams) (database.GetMonetaryDataForMonthRow, error) {
	if m.GetMonetaryDataForMonthFunc != nil {
		return m.GetMonetaryDataForMonthFunc(ctx, arg)
	}
	return database.GetMonetaryDataForMonthRow{}, nil
}

func (m *mockDatabaseService) ValidateCurrency(ctx context.Context, code string) (bool, error) {
	if m.ValidateCurrencyFunc != nil {
		return m.ValidateCurrencyFunc(ctx, code)
	}
	return false, nil
}

func (m *mockDatabaseService) CreateDelegation(ctx context.Context, arg database.CreateDelegationParams) (database.Delegation, error) {
	if m.CreateDelegationFunc != nil {
		return m.CreateDelegationFunc(ctx, arg)
	}
	return database.Delegation{}, nil
}

func (m *mockDatabaseService) GetDelegation(ctx context.Context, id uuid.UUID) (database.Delegation, error) {
	if m.GetDelegationFunc != nil {
		return m.GetDelegationFunc(ctx, id)
	}
	return database.Delegation{}, nil
}

func (m *mockDatabaseService) RevokeDelegationByID(ctx context.Context, id uuid.UUID) error {
	if m.RevokeDelegationByIDFunc != nil {
		return m.RevokeDelegationByIDFunc(ctx, id)
	}
	return nil
}

func (m *mockDatabaseService) RevokeDelegationByUser(ctx context.Context, userID uuid.UUID) error {
	if m.RevokeDelegationByUserFunc != nil {
		return m.RevokeDelegationByUserFunc(ctx, userID)
	}
	return nil
}

func (m *mockDatabaseService) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	if m.UpdateLastUsedFunc != nil {
		return m.UpdateLastUsedFunc(ctx, id)
	}
	return nil
}

func (m *mockDatabaseService) CreateItem(ctx context.Context, arg database.CreateItemParams) (database.PlaidItem, error) {
	if m.CreateItemFunc != nil {
		return m.CreateItemFunc(ctx, arg)
	}
	return database.PlaidItem{}, nil
}

func (m *mockDatabaseService) DeleteItem(ctx context.Context, arg database.DeleteItemParams) error {
	if m.DeleteItemFunc != nil {
		return m.DeleteItemFunc(ctx, arg)
	}
	return nil
}

func (m *mockDatabaseService) GetAccessToken(ctx context.Context, id string) (database.PlaidItem, error) {
	if m.GetAccessTokenFunc != nil {
		return m.GetAccessTokenFunc(ctx, id)
	}
	return database.PlaidItem{}, nil
}

func (m *mockDatabaseService) GetCursor(ctx context.Context, arg database.GetCursorParams) (sql.NullString, error) {
	if m.GetCursorFunc != nil {
		return m.GetCursorFunc(ctx, arg)
	}
	return sql.NullString{}, nil
}

func (m *mockDatabaseService) GetItemByID(ctx context.Context, id string) (database.PlaidItem, error) {
	if m.GetItemByIDFunc != nil {
		return m.GetItemByIDFunc(ctx, id)
	}
	return database.PlaidItem{}, nil
}

func (m *mockDatabaseService) GetItemByName(ctx context.Context, nickname sql.NullString) (database.PlaidItem, error) {
	if m.GetItemByNameFunc != nil {
		return m.GetItemByNameFunc(ctx, nickname)
	}
	return database.PlaidItem{}, nil
}

func (m *mockDatabaseService) GetItemsByUser(ctx context.Context, userID uuid.UUID) ([]database.PlaidItem, error) {
	if m.GetItemsByUserFunc != nil {
		return m.GetItemsByUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockDatabaseService) GetLatestCursorOrNil(ctx context.Context, id string) (sql.NullString, error) {
	if m.GetLatestCursorOrNilFunc != nil {
		return m.GetLatestCursorOrNilFunc(ctx, id)
	}
	return sql.NullString{}, nil
}

func (m *mockDatabaseService) ResetItems(ctx context.Context) error {
	if m.ResetItemsFunc != nil {
		return m.ResetItemsFunc(ctx)
	}
	return nil
}

func (m *mockDatabaseService) UpdateCursor(ctx context.Context, arg database.UpdateCursorParams) error {
	if m.UpdateCursorFunc != nil {
		return m.UpdateCursorFunc(ctx, arg)
	}
	return nil
}

func (m *mockDatabaseService) UpdateNickname(ctx context.Context, arg database.UpdateNicknameParams) error {
	if m.UpdateNicknameFunc != nil {
		return m.UpdateNicknameFunc(ctx, arg)
	}
	return nil
}

func (m *mockDatabaseService) CreateToken(ctx context.Context, arg database.CreateTokenParams) (database.RefreshToken, error) {
	if m.CreateTokenFunc != nil {
		return m.CreateTokenFunc(ctx, arg)
	}
	return database.RefreshToken{}, nil
}

func (m *mockDatabaseService) ExpireAllDelegationTokens(ctx context.Context, delegationID uuid.UUID) error {
	if m.ExpireAllDelegationTokensFunc != nil {
		return m.ExpireAllDelegationTokensFunc(ctx, delegationID)
	}
	return nil
}

func (m *mockDatabaseService) ExpireToken(ctx context.Context, hashedToken string) error {
	if m.ExpireTokenFunc != nil {
		return m.ExpireTokenFunc(ctx, hashedToken)
	}
	return nil
}

func (m *mockDatabaseService) GetToken(ctx context.Context, hashedToken string) (database.RefreshToken, error) {
	if m.GetTokenFunc != nil {
		return m.GetTokenFunc(ctx, hashedToken)
	}
	return database.RefreshToken{}, nil
}

func (m *mockDatabaseService) ClearTransactionsTable(ctx context.Context) error {
	if m.ClearTransactionsTableFunc != nil {
		return m.ClearTransactionsTableFunc(ctx)
	}
	return nil
}

func (m *mockDatabaseService) CreateTransaction(ctx context.Context, arg database.CreateTransactionParams) (database.Transaction, error) {
	if m.CreateTransactionFunc != nil {
		return m.CreateTransactionFunc(ctx, arg)
	}
	return database.Transaction{}, nil
}

func (m *mockDatabaseService) DeleteTransaction(ctx context.Context, arg database.DeleteTransactionParams) error {
	if m.DeleteTransactionFunc != nil {
		return m.DeleteTransactionFunc(ctx, arg)
	}
	return nil
}

func (m *mockDatabaseService) DeleteTransactionsForAccount(ctx context.Context, accountID string) error {
	if m.DeleteTransactionsForAccountFunc != nil {
		return m.DeleteTransactionsForAccountFunc(ctx, accountID)
	}
	return nil
}

func (m *mockDatabaseService) GetTransactions(ctx context.Context, accountID string) ([]database.Transaction, error) {
	if m.GetTransactionsFunc != nil {
		return m.GetTransactionsFunc(ctx, accountID)
	}
	return nil, nil
}

func (m *mockDatabaseService) GetTransactionsForUser(ctx context.Context, userID uuid.UUID) ([]database.Transaction, error) {
	if m.GetTransactionsForUserFunc != nil {
		return m.GetTransactionsForUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockDatabaseService) CreateVerificationRecord(ctx context.Context, arg database.CreateVerificationRecordParams) (database.VerificationRecord, error) {
	if m.CreateVerificationRecordFunc != nil {
		return m.CreateVerificationRecordFunc(ctx, arg)
	}
	return database.VerificationRecord{}, nil
}

func (m *mockDatabaseService) DeleteVerificationRecord(ctx context.Context, verificationCode string) error {
	if m.DeleteVerificationRecordFunc != nil {
		return m.DeleteVerificationRecordFunc(ctx, verificationCode)
	}
	return nil
}

func (m *mockDatabaseService) DeleteVerificationRecordByUser(ctx context.Context, userID uuid.UUID) error {
	if m.DeleteVerificationRecordByUserFunc != nil {
		return m.DeleteVerificationRecordByUserFunc(ctx, userID)
	}
	return nil
}

func (m *mockDatabaseService) GetVerificationRecord(ctx context.Context, verificationCode string) (database.VerificationRecord, error) {
	if m.GetVerificationRecordFunc != nil {
		return m.GetVerificationRecordFunc(ctx, verificationCode)
	}
	return database.VerificationRecord{}, nil
}

func (m *mockDatabaseService) GetVerificationRecordByUser(ctx context.Context, userID uuid.UUID) (database.VerificationRecord, error) {
	if m.GetVerificationRecordByUserFunc != nil {
		return m.GetVerificationRecordByUserFunc(ctx, userID)
	}
	return database.VerificationRecord{}, nil
}

func (m *mockDatabaseService) CreatePlaidWebhookRecord(ctx context.Context, arg database.CreatePlaidWebhookRecordParams) (database.PlaidWebhookRecord, error) {
	if m.CreatePlaidWebhookRecordFunc != nil {
		return m.CreatePlaidWebhookRecordFunc(ctx, arg)
	}
	return database.PlaidWebhookRecord{}, nil
}

func (m *mockDatabaseService) ProcessWebhookRecordsByType(ctx context.Context, arg database.ProcessWebhookRecordsByTypeParams) error {
	if m.ProcessWebhookRecordsByTypeFunc != nil {
		return m.ProcessWebhookRecordsByTypeFunc(ctx, arg)
	}
	return nil
}

func (m *mockDatabaseService) GetWebhookRecords(ctx context.Context, userID uuid.UUID) ([]database.PlaidWebhookRecord, error) {
	if m.GetWebhookRecordsFunc != nil {
		return m.GetWebhookRecordsFunc(ctx, userID)
	}
	return []database.PlaidWebhookRecord{}, nil
}

func (m *mockAuthService) EmailValidation(email string) bool {
	if m.EmailValidationFunc != nil {
		return m.EmailValidationFunc(email)
	}
	return true
}

func (m *mockAuthService) HashPassword(password string) (string, error) {
	if m.HashPasswordFunc != nil {
		return m.HashPasswordFunc(password)
	}
	return password + "_hashed", nil
}

func (m *mockAuthService) ValidatePasswordHash(hash, password string) error {
	if m.ValidatePasswordHashFunc != nil {
		return m.ValidatePasswordHashFunc(hash, password)
	}
	return nil
}

func (m *mockAuthService) GenerateCode() string {
	if m.GenerateCodeFunc != nil {
		return m.GenerateCodeFunc()
	}
	return "12345"
}

func (m *mockAuthService) GetBearerToken(headers http.Header) (string, error) {
	if m.GetBearerTokenFunc != nil {
		return m.GetBearerTokenFunc(headers)
	}
	return "token", nil
}

func (m *mockAuthService) MakeJWT(cfg *config.Config, userID uuid.UUID) (string, error) {
	if m.MakeJWTFunc != nil {
		return m.MakeJWTFunc(cfg, userID)
	}
	return "JWTtoken", nil
}

func (m *mockAuthService) ValidateJWT(cfg *config.Config, tokenString string) (uuid.UUID, error) {
	if m.ValidateJWTFunc != nil {
		return m.ValidateJWTFunc(cfg, tokenString)
	}
	return uuid.UUID{}, nil
}

func (m *mockAuthService) VerifyPlaidJWT(p auth.PlaidKeyFetcher, ctx context.Context, tokenString string) error {
	if m.VerifyPlaidJWTFunc != nil {
		return m.VerifyPlaidJWTFunc(p, ctx, tokenString)
	}
	return nil
}

func (m *mockAuthService) MakeRefreshToken(tokenStore auth.TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
	if m.MakeRefreshTokenFunc != nil {
		return m.MakeRefreshTokenFunc(tokenStore, userID, delegation)
	}
	return "refreshToken", nil
}

func (m *mockAuthService) HashRefreshToken(token string) string {
	if m.HashRefreshTokenFunc != nil {
		return m.HashRefreshTokenFunc(token)
	}
	return token + "_hash"
}

func (m *mockMailService) NewMail(from string, to string, subject string, body string, data *sgrid.MailData) *sgrid.Mail {
	if m.NewMailFunc != nil {
		return m.NewMailFunc(from, to, subject, body, data)
	}
	return &sgrid.Mail{}
}

func (m *mockMailService) SendMail(mailreq *sgrid.Mail) error {
	if m.SendMailFunc != nil {
		return m.SendMailFunc(mailreq)
	}
	return nil
}

func (p *mockPlaidService) GetLinkToken(ctx context.Context, userID, webhookURL string) (string, error) {
	if p.GetLinkTokenFunc != nil {
		return p.GetLinkTokenFunc(ctx, userID, webhookURL)
	}
	return "", nil
}

func (p *mockPlaidService) GetLinkTokenForUpdateMode(ctx context.Context, userID, accessToken, webhookURL string) (string, error) {
	if p.GetLinkTokenForUpdateModeFunc != nil {
		return p.GetLinkTokenForUpdateModeFunc(ctx, userID, accessToken, webhookURL)
	}
	return "", nil
}

func (p *mockPlaidService) GetAccessToken(ctx context.Context, publicToken string) (models.AccessResponse, error) {
	if p.GetAccessTokenFunc != nil {
		return p.GetAccessTokenFunc(ctx, publicToken)
	}
	return models.AccessResponse{}, nil
}

func (p *mockPlaidService) InvalidateAccessToken(ctx context.Context, accessToken models.AccessResponse) (models.AccessResponse, error) {
	if p.InvalidateAccessTokenFunc != nil {
		return p.InvalidateAccessTokenFunc(ctx, accessToken)
	}
	return models.AccessResponse{}, nil
}

func (p *mockPlaidService) GetAccounts(ctx context.Context, accessToken string) ([]plaid.AccountBase, string, error) {
	if p.GetAccountsFunc != nil {
		return p.GetAccountsFunc(ctx, accessToken)
	}
	return []plaid.AccountBase{}, "", nil
}

func (p *mockPlaidService) GetItemInstitution(ctx context.Context, accessToken string) (string, error) {
	if p.GetItemInstitutionFunc != nil {
		return p.GetItemInstitutionFunc(ctx, accessToken)
	}
	return "TestInstitution", nil
}

func (p *mockPlaidService) GetBalances(ctx context.Context, accessToken string) (plaid.AccountsGetResponse, string, error) {
	if p.GetBalancesFunc != nil {
		return p.GetBalancesFunc(ctx, accessToken)
	}
	return plaid.AccountsGetResponse{}, "", nil
}

func (p *mockPlaidService) GetTransactions(ctx context.Context, accessToken, cursor string) (
	added, modified []plaid.Transaction, removed []plaid.RemovedTransaction, nextCursor, reqID string, err error,
) {
	if p.GetTransactionsFunc != nil {
		return p.GetTransactionsFunc(ctx, accessToken, cursor)
	}
	return []plaid.Transaction{}, []plaid.Transaction{}, []plaid.RemovedTransaction{}, "next", "request", nil
}

func (p *mockPlaidService) GetWebhookVerificationKey(ctx context.Context, keyID string) (plaid.JWKPublicKey, error) {
	if p.GetWebhookVerificationKeyFunc != nil {
		return p.GetWebhookVerificationKeyFunc(ctx, keyID)
	}
	return plaid.JWKPublicKey{}, nil
}

func (p *mockPlaidService) RemoveItem(ctx context.Context, accessToken string) error {
	if p.RemoveItemFunc != nil {
		return p.RemoveItemFunc(ctx, accessToken)
	}
	return nil
}

func (t *mockTxnUpdaterService) ExpireDelegation(ctx context.Context, tokenHash string, token database.RefreshToken) error {
	if t.ExpireDelegationFunc != nil {
		return t.ExpireDelegationFunc(ctx, tokenHash, token)
	}
	return nil
}

func (t *mockTxnUpdaterService) RevokeDelegation(ctx context.Context, token database.RefreshToken) error {
	if t.RevokeDelegationFunc != nil {
		return t.RevokeDelegationFunc(ctx, token)
	}
	return nil
}

func (t *mockTxnUpdaterService) ApplyTransactionUpdates(ctx context.Context, added []plaid.Transaction, modified []plaid.Transaction, removed []plaid.RemovedTransaction, nextCursor, itemID string) error {
	if t.ApplyTransactionUpdatesFunc != nil {
		return t.ApplyTransactionUpdatesFunc(ctx, added, modified, removed, nextCursor, itemID)
	}
	return nil
}

func (e *mockEncryptor) EncryptAccessToken(plaintext []byte, keyString string) (string, error) {
	if e.EncryptAccessTokenFunc != nil {
		return e.EncryptAccessTokenFunc(plaintext, keyString)
	}
	return "encrypted", nil
}

func (e *mockEncryptor) DecryptAccessToken(ciphertext, keyString string) ([]byte, error) {
	if e.DecryptAccessTokenFunc != nil {
		return e.DecryptAccessTokenFunc(ciphertext, keyString)
	}
	return []byte{123}, nil
}

func (q *mockQuerier) ValidateParamValue(value, expectedType string) (bool, error) {
	if q.ValidateParamValueFunc != nil {
		return q.ValidateParamValueFunc(value, expectedType)
	}
	return true, nil
}

func (q *mockQuerier) ValidateQuery(queries url.Values, rules map[string]string) (map[string]string, []utils.QueryValidationError) {
	if q.ValidateQueryFunc != nil {
		return q.ValidateQueryFunc(queries, rules)
	}
	return map[string]string{}, []utils.QueryValidationError{}
}

func (q *mockQuerier) BuildSqlQuery(queries map[string]string, accountID string) (string, []any, error) {
	if q.BuildSqlQueryFunc != nil {
		return q.BuildSqlQueryFunc(queries, accountID)
	}
	return "", nil, nil
}
