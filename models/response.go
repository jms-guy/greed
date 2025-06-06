package models

import (
	"time"
	"github.com/google/uuid"
)

//Token refresh response structure
type RefreshResponse struct{
	RefreshToken	string	`json:"refresh_token"`
	AccessToken		string	`json:"access_token"`
	TokenType		string	`json:"token_type"`
}

//Login response structure
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
	Email			string		`json:"email"`
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

//Income structure 
type Income struct {
	Amount			string 	`json:"amount"`
	Year			int 	`json:"year"`
	Month			int 	`json:"month"`
}

//Expense structure
type Expenses struct {
	Amount			string 	`json:"amount"`
	Year			int 	`json:"year"`
	Month			int 	`json:"month"`
}

//Verification email response structure
type VerificationData struct {
	Email		string 					`json:"email"`
	Code 		string					`json:"code"`
	ExpiresAt	time.Time				`json:"expires_at"`
}
