package main

import (
	"time"
	"github.com/google/uuid"
)

//Structure for account JSON response 
type Account struct {
	ID				uuid.UUID `json:"id"`
	CreatedAt		time.Time `json:"created_at"`
	UpdatedAt		time.Time `json:"updated_at"`
	Balance			string `json:"balance"`
	Goal			string `json:"goal"`
	Currency		string `json:"currency"`
	UserID			uuid.UUID `json:"user_id"`
}

//Structure for user JSON response
type User struct{
	ID				uuid.UUID `json:"id"`
	Name			string `json:"name"`
	CreatedAt		time.Time `json:"created_at"`
	UpdatedAt		time.Time `json:"updated_at"`	
}

