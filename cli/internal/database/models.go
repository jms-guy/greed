// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"database/sql"
)

type Account struct {
	ID               string
	CreatedAt        string
	UpdatedAt        string
	Name             string
	Type             string
	Subtype          sql.NullString
	Mask             sql.NullString
	OfficialName     sql.NullString
	AvailableBalance sql.NullFloat64
	CurrentBalance   sql.NullFloat64
	IsoCurrencyCode  sql.NullString
	InstitutionName  sql.NullString
	UserID           string
}

type Transaction struct {
	ID                      string
	AccountID               string
	Amount                  float64
	IsoCurrencyCode         sql.NullString
	Date                    sql.NullString
	MerchantName            sql.NullString
	PaymentChannel          string
	PersonalFinanceCategory string
}

type User struct {
	ID             string
	Name           string
	CreatedAt      string
	UpdatedAt      string
	HashedPassword string
	Email          string
	IsVerified     sql.NullBool
}
