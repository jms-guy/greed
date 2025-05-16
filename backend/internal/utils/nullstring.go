package utils

import (
	"database/sql"
	"fmt"
)

//Turns a given string into SQL type NullString, so it can be
//inserted into possible Null fields in database - this version
//handles strings that should be monetary values
func CreateMoneyNullString(s string) (sql.NullString, error) {
	nullString := sql.NullString{}

	if s == "" {
		nullString.Valid = false
		//Validates string, making sure it is in proper format
	} else if moneyStringValidation(s) {
		nullString.Valid = true
		nullString.String = s
	} else {
		return sql.NullString{}, fmt.Errorf("invalid string format, need (xxx.xx)")
	}
	return nullString, nil
}

//This version will handle text NullStrings
func CreateTextNullString(s string) sql.NullString {
	nullString := sql.NullString{}

	if s == "" {
		nullString.Valid = false
	} else {
		nullString.Valid = true
		nullString.String = s
	}
	return nullString
}