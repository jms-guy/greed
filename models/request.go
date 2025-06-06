package models

import (
	"time"
	"github.com/google/uuid"
)

type RefreshRequest struct {
	RefreshToken 			string 	`json:"refresh_token"`
}

//Request parameters for users
type UserDetails struct {
	Name					string `json:"name"`
	Password				string `json:"password"`	
	Email					string `json:"email"`
}

//Request struct for email verification - with code
type EmailVerificationWithCode struct {
	UserID 					uuid.UUID 	`json:"user_id"`
	Code 					string 		`json:"code"`
}

//Request struct for resetting user's forgotten password
type ResetPassword struct {
	Email 					string 		`json:"email"`
	Code 					string		`json:"code"`		//email verification code
	NewPassword				string 		`json:"new_password"`
}

//Request struct for updating user's password
type UpdatePassword struct {
	NewPassword				string 		`json:"new_password"`
	Code					string 		`json:"code"`		//email verification code
}

//Request struct for email verification - without code
type EmailVerification struct {
	UserID 					uuid.UUID	`json:"user_id"`
	Email 					string 		`json:"email"`
}

//Request parameters for creating an account
type AccountDetails struct {
	Name					string `json:"name"`
}

//Request parameters needed for creating a transaction record
type CreateTransaction struct {
	Amount						string 	  `json:"amount"`
	Category					string	  `json:"category"`
	Description					string	  `json:"description"`
	TransactionDate 			time.Time `json:"transaction_date"`
	TransactionType 			string	  `json:"transaction_type"`
	CurrencyCode				string	  `json:"currency_code"`
	AccountID					uuid.UUID `json:"account_id"`
}

//Request parameters for updating a transaction description
type UpdateDescription struct {
	Description		string `json:"description"`
}

//Parameters for updating a transaction category
type UpdateCategory struct {
	Category		string `json:"category"`
}

/*
//Parameters for updating an account's set currency
type UpdateCurrency struct {
	Currency		string `json:"currency"`
}

//Parameters for updating an account's goal
type UpdateGoal struct {
	Goal		string `json:"goal"`
}

//Parameters for updating an account's balance
type UpdateBalance struct {
	Balance		string`json:"balance"`
}
*/