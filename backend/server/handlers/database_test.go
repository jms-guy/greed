package handlers_test

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
)

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