package models

import (
	"time"
	"github.com/google/uuid"
)

//Request parameters for users
type UserDetails struct {
	Name			string `json:"name"`
	Password		string `json:"password"`	
}

//Request parameters for creating an account
type CreateAccount struct {
	Name		string `json:"name"`
	Currency	string `json:"currency"`
	UserID		uuid.UUID `json:"user_id"`
}

//Request parameters needed for creating a transaction record
type CreateTransaction struct {
	Amount			string 	  `json:"amount"`
	Category		string	  `json:"category"`
	Description		string	  `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
	TransactionType string	  `json:"transaction_type"`
	CurrencyCode	string	  `json:"currency_code"`
	AccountID		uuid.UUID `json:"account_id"`
}

//Request parameters for updating a transaction description
type UpdateDescription struct {
	Description		string `json:"description"`
}

//Parameters for updating a transaction category
type UpdateCategory struct {
	Category		string `json:"category"`
}

//Parameters for updating an account's set currency
type UpdateCurrency struct {
	Currency		string `json:"currency"`
}

/*
//Parameters for updating an account's goal
type UpdateGoal struct {
	Goal		string `json:"goal"`
}

//Parameters for updating an account's balance
type UpdateBalance struct {
	Balance		string`json:"balance"`
}
*/