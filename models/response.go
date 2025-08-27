package models

import (
	"time"

	"github.com/google/uuid"
)

type ItemName struct {
	Nickname        string `json:"nickname"`
	ItemId          string `json:"item_id"`
	InstitutionName string `json:"institution_name"`
}

type Accounts struct {
	Accounts  []Account `json:"accounts"`
	RequestID string    `json:"request_id"` // This field is the request ID returned from the Plaid API call
}

type Account struct {
	Id               string    `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Name             string    `json:"name:"`
	Type             string    `json:"type"`
	Subtype          string    `json:"subtype"`
	Mask             string    `json:"mask"`
	OfficialName     string    `json:"official_name"`
	AvailableBalance string    `json:"available_balance"`
	CurrentBalance   string    `json:"current_balance"`
	IsoCurrencyCode  string    `json:"iso_currency_code"`
	ItemId           string    `json:"item_id"`
}

type Transaction struct {
	Id                      string    `json:"id"`
	AccountId               string    `json:"account_id"`
	Amount                  string    `json:"amount"`
	IsoCurrencyCode         string    `json:"iso_currency_code"`
	Date                    time.Time `json:"date"`
	MerchantName            string    `json:"merchant_name"`
	PaymentChannel          string    `json:"payment_channel"`
	PersonalFinanceCategory string    `json:"personal_finance_category"`
}

type UpdatedBalance struct {
	Id               string `json:"id"`
	AvailableBalance string `json:"available_balance"`
	CurrentBalance   string `json:"current_balance"`
	ItemId           string `json:"item_id"`
	RequestID        string `json:"request_id"`
}

type AccessResponse struct {
	AccessToken     string `json:"access_token"`
	ItemID          string `json:"item_id"`
	InstitutionName string `json:"institution_name"`
	RequestID       string `json:"request_id"` // Request ID returned from Plaid API call
}

type LinkResponse struct {
	LinkToken string `json:"link_token"`
}

type UpdatedPassword struct {
	HashPassword string `json:"hash_password"`
}

type RefreshResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
}

type Credentials struct {
	User         User   `json:"user"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type User struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"hashed_password"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// For credit/debit account types, fields are straightforward. For loan accounts, income =
// loan payments, expenses = interest, netincome = income - expenses
type MonetaryData struct {
	Income    string `json:"income"`
	Expenses  string `json:"expenses"`
	NetIncome string `json:"net_income"`
	Date      string `json:"date"`
}

// For getting merchant summaries on transactions call
type MerchantSummary struct {
	Merchant    string `json:"merchant"`
	TxnCount    int64  `json:"txn_count"`
	Category    string `json:"category"`
	TotalAmount string `json:"total_amount"`
	Month       string `json:"month"`
}

type WebhookRecord struct {
	WebhookType string    `json:"webhook_type"`
	WebhookCode string    `json:"webhook_code"`
	ItemID      string    `json:"item_id"`
	UserID      uuid.UUID `json:"user_id"`
	CreatedAt   string    `json:"created_at"`
}

type RecurringData struct {
	Streams     []RecurringStream      `json:"streams"`
	Connections []TransactionsToStream `json:"connections"`
}

type RecurringStream struct {
	ID                string `json:"id"`
	AccountID         string `json:"account_id"`
	Description       string `json:"description"`
	MerchantName      string `json:"merchant_name"`
	Frequency         string `json:"frequency"`
	IsActive          bool   `json:"is_active"`
	PredictedNextDate string `json:"predicted_next_date"`
	StreamType        string `json:"stream_type"`
}

type TransactionsToStream struct {
	TransactionID string `json:"transaction_id"`
	StreamID      string `json:"stream_id"`
}
