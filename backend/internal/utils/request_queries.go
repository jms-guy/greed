package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//Represents an error related to query parameter validation
type QueryValidationError struct {
	Parameter 		string 	`json:"parameter"`
	Value 			string 	`json:"value"`
	Message 		string 	`json:"message"`
}

//Responsible for validating query parameters based on their type. 
//Holds map of type validators, where key is the type, and value is a function that validates the string value
type QueryValidator struct {
	typeValidators 	map[string]func(string) bool 
}

//Initializes new QueryValidator instance
func NewQueryValidator() *QueryValidator {
	qv := &QueryValidator{
		typeValidators: make(map[string]func(string) bool),
	}

	qv.typeValidators["number"] = func(v string) bool {
		matched, _ := regexp.MatchString(`^-?\d+(\.\d+)?$`, v)	//Validator for numerics (int and float)
		return matched
	}

	qv.typeValidators["boolean"] = func(v string) bool {
		v = strings.ToLower(v)
		return v == "true" || v == "false" || v == "1" || v == "0" //Validator for boolean
	}

	qv.typeValidators["order"] = func(v string) bool {
		v = strings.ToLower(v)
		return v == "asc" || v == "desc"			//Validator for query ordering
	}

	qv.typeValidators["date"] = func(v string) bool {
		_, err := time.Parse("2006-01-02", v)
		return err == nil
	}

	return qv 
}

//Validates a query parameter's value based on its expected type.
// - value: Value of parameter to validate
// - expectedType: Expected data type of query parameter
// Returns true if value passes validation or no validator exists for expected type
func (qv *QueryValidator) validateParamValue(value, expectedType string) bool {
	validator, exists := qv.typeValidators[expectedType]
	if !exists {
		return true
	}

	return validator(value)
}

//Validates query parameters based on predefined rules. 
//Rules is a map of expected query parameters and their expected types. 
//Returns slice of QueryValidationError if validation fails
func (qv *QueryValidator) ValidateQuery(queries url.Values, rules map[string]string) (map[string]string, []QueryValidationError) {
	var errors []QueryValidationError

	var parsed = make(map[string]string)

	for key, value := range queries {
		parsed[key] = value[0]
	}

	for param, value := range parsed {
		expectedType, exists := rules[param]
		if !exists {
			errors = append(errors, QueryValidationError{
				Parameter: param,
				Value: value,
				Message: "unexpected parameter",
			})
			continue
		}

		if !qv.validateParamValue(value, expectedType) {
			errors = append(errors, QueryValidationError{
				Parameter: param,
				Value: value,
				Message: fmt.Sprintf("invalid value for type %s", expectedType),
			})
		}
	}

	return parsed, errors
}

//Builds an SQL query for transactions based on optional query arguments
func BuildSqlQuery(queries map[string]string, accountID string) (string, []any, error) {
	query := "SELECT * FROM transactions WHERE account_id = $1"
	args := []any{accountID}
	paramCount := 2

	if val, ok := queries["merchant"]; ok {
		if val != "" {
			query += fmt.Sprintf(" AND merchant_name = $%d", paramCount)
			args = append(args, val)
			paramCount++
		}
	}
	if val, ok := queries["category"]; ok {
		if val != "" {
			query += fmt.Sprintf(" AND personal_finance_category = $%d", paramCount)
			args = append(args, val)
			paramCount++
		}
	}
	if val, ok := queries["channel"]; ok {
		if val != "" {
			query += fmt.Sprintf(" AND payment_channel = $%d", paramCount)
			args = append(args, val)
			paramCount++
		}
	}
	if val, ok := queries["date"]; ok {
		_, err := time.Parse("2006-01-02", val)
		if err != nil {
			return "", args, err 
		}
		query += fmt.Sprintf(" AND DATE(date) = $%d", paramCount)
		args = append(args, val)
		paramCount++
	} else {
		if start, exists := queries["start"]; exists {
			_, err := time.Parse("2006-01-02", start)
			if err != nil {
				return "", args, err 
			}
			query += fmt.Sprintf(" AND DATE(date) >= $%d", paramCount)
			args = append(args, start)
			paramCount++
		}
		if end, exists := queries["end"]; exists {
			_, err := time.Parse("2006-01-02", end)
			if err != nil {
				return "", args, err
			}
			query += fmt.Sprintf(" AND DATE(date) <= $%d", paramCount)
			args = append(args, end)
			paramCount++
		}
	}
	if val, ok := queries["min"]; ok {
		if val != "" {
			fVal, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return "", args, err
			}
			query += fmt.Sprintf(" AND amount >= $%d", paramCount)
			args = append(args, fVal)
			paramCount++
		}
	}
	if val, ok := queries["max"]; ok {
		if val != "" {
			fVal, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return "", args, err
			}
			query += fmt.Sprintf(" AND amount <= $%d", paramCount)
			args = append(args, fVal)
			paramCount++
		}
	}
	if val, ok := queries["order"]; ok {
		if val != "" {
			order := "DESC" // default
			if val, ok := queries["order"]; ok && (strings.ToLower(val) == "asc" || strings.ToLower(val) == "desc") {
				order = strings.ToUpper(val)
			}
			query += fmt.Sprintf(" ORDER BY date %s", order)
		}
	} else {
		query += " ORDER BY date DESC"
	}
	if val, ok := queries["limit"]; ok {
		if val != "" {
			limit, err := strconv.Atoi(val)
			if err != nil {
				return "", args, err
			}
			if limit > 200 {limit = 200}
			query += fmt.Sprintf(" LIMIT $%d", paramCount)
			args = append(args, limit)
			paramCount++
		}
	}

	return query, args, nil
}