package models

import (
	"time"
	"github.com/google/uuid"
)

//Structs used in http responses

/*
	/api/items
*/
type ItemName struct {
	Nickname		string 	`json:"nickname"`
	ItemId 			string 	`json:"item_id"`
	InstitutionName string 	`json:"institution_name"`
}

/*
	/api/items/accounts
*/
type Accounts struct {
	Accounts 		[]Account`json:"accounts"`
	RequestID		string 	`json:"request_id"`	//This field is the request ID returned from the Plaid API call
}

/*
	/api/items/accounts
	/api/accounts
	/api/accounts/{item-id}
	/api/accounts/{account-id}/data
*/
type Account struct {
	Id 				string 	`json:"id"`
	CreatedAt 		time.Time 	`json:"created_at"`
	UpdatedAt 		time.Time 	`json:"updated_at"`
	Name			string 	`json:"name:"`
	Type 			string  `json:"type"`
	Subtype 		string  `json:"subtype"`
	Mask 			string  `json:"mask"`
	OfficialName    string  `json:"official_name"`
	AvailableBalance string `json:"available_balance"`
	CurrentBalance  string  `json:"current_balance"`
	IsoCurrencyCode string  `json:"iso_currency_code"`
	ItemId 			string 	`json:"item_id"`
}

/*
	/api/accounts/{account-id}/transactions
*/
type Transaction struct {
	Id 							string 		`json:"id"`
	AccountId 					string 		`json:"account_id"`
	Amount 						string 		`json:"amount"`
	IsoCurrencyCode 			string 		`json:"iso_currency_code"`
	Date 						time.Time 	`json:"date"`
	MerchantName 				string 		`json:"merchant_name"`
	PaymentChannel				string 		`json:"payment_channel"`
	PersonalFinanceCategory 	string 		`json:"personal_finance_category"`
}

/*
	/api/accounts/balance
*/
type UpdatedBalance struct {
	Id 				string 	`json:"id"`
	AvailableBalance string `json:"available_balance"`
	CurrentBalance  string 	`json:"current_balance"`
	ItemId			string 	`json:"item_id"`
	RequestID 		string 	`json:"request_id"`
}

/*
	/plaid/get-access-token
*/
type AccessResponse struct {
	AccessToken 	string 	`json:"access_token"`
	ItemID			string  `json:"item_id"`
	InstitutionName string 	`json:"institution_name"`
	RequestID 		string  `json:"request_id"`	//Request ID returned from Plaid API call
}

/*
	/plaid/get-link-token
*/
type LinkResponse struct {
	LinkToken		string 	`json:"link_token"`
}

/*
	/auth/refresh
*/
type RefreshResponse struct{
	RefreshToken	string	`json:"refresh_token"`
	AccessToken		string	`json:"access_token"`
	TokenType		string	`json:"token_type"`
}

/*
	/auth/login
*/
type LoginResponse struct{
	User 			User	`json:"user"`
	RefreshToken	string	`json:"refresh_token"`
	AccessToken		string	`json:"access_token"`
	TokenType		string	`json:"token_type"`
	ExpiresIn		int		`json:"expires_in"`
}

/*
	/auth/register
*/
type User struct{
	ID				uuid.UUID 	`json:"id"`
	Name			string 		`json:"name"`
	Email			string		`json:"email"`
	HashedPassword  string 		`json:"hashed_password"`
	CreatedAt		time.Time 	`json:"created_at"`
	UpdatedAt		time.Time 	`json:"updated_at"`	
}

/*
	/api/accounts/{account-id}/transactions/income
	/api/accounts/{account-id}/transactions/income/{year}-{month}
*/
//For credit/debit account types, fields are straightforward. For loan accounts, income =  
//loan payments, expenses = interest, netincome = income - expenses
type MonetaryData struct {
	Income 			string 		`json:"income"`
	Expenses 		string 		`json:"expenses"`
	NetIncome 		string 		`json:"net_income"`
	Year 			int 		`json:"year"`
	Month 			int 		`json:"month"`
}


