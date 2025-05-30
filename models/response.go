package models

import (
	"time"
	"github.com/google/uuid"
)

type RefreshResponse struct{
	RefreshToken	string	`json:"refresh_token"`
	AccessToken		string	`json:"access_token"`
	TokenType		string	`json:"token_type"`
	ExpiresIn		int		`json:"expires_in"`
}

type LoginResponse struct{
	User 			User	`json:"user"`
	RefreshToken	string	`json:"refresh_token"`
	AccessToken		string	`json:"access_token"`
	TokenType		string	`json:"token_type"`
	ExpiresIn		int		`json:"expires_in"`
}

//Structure for account JSON response 
type Account struct {
	ID				uuid.UUID 	`json:"id"`
	CreatedAt		time.Time 	`json:"created_at"`
	UpdatedAt		time.Time 	`json:"updated_at"`
	Name			string		`json:"name"`
	UserID			uuid.UUID 	`json:"user_id"`
}

//Structure for user JSON response
type User struct{
	ID				uuid.UUID 	`json:"id"`
	Name			string 		`json:"name"`
	CreatedAt		time.Time 	`json:"created_at"`
	UpdatedAt		time.Time 	`json:"updated_at"`	
}

//Structure for transaction JSON response
type Transaction struct {
	ID				uuid.UUID `json:"id"`
	CreatedAt		time.Time `json:"created_at"`
	UpdatedAt		time.Time `json:"updated_at"`
	Amount			string 	  `json:"amount"`
	Category		string	  `json:"category"`
	Description		string	  `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
	TransactionType string	  `json:"transaction_type"`
	CurrencyCode	string	  `json:"currency_code"`
	AccountID		uuid.UUID `json:"account_id"`
}

//Income structure - used for income/expenses
type Income struct {
	Amount			string 	`json:"amount"`
	Year			int 	`json:"year"`
	Month			int 	`json:"month"`
}

type Expenses struct {
	Amount			string 	`json:"amount"`
	Year			int 	`json:"year"`
	Month			int 	`json:"month"`
}
