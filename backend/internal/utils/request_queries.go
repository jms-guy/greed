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
type Service struct {
	typeValidators 	map[string]func(string) bool 
}

//URL Query service interface
type QueryService interface {
	ValidateParamValue(value, expectedType string) (bool, error)
	ValidateQuery(queries url.Values, rules map[string]string) (map[string]string, []QueryValidationError)
	BuildSqlQuery(queries map[string]string, accountID string) (string, []any, error)
}

//Initializes new QueryValidator instance
func NewQueryService() *Service {
	qv := &Service{
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
func (qv *Service) ValidateParamValue(value, expectedType string) (bool, error) {
	validator, exists := qv.typeValidators[expectedType]
	if !exists {
		return true, nil
	}

	if strings.Contains(value, ";") {
		return false, fmt.Errorf("query value contains ';' character")
	}

	return validator(value), nil
}

//Validates query parameters based on predefined rules. 
//Rules is a map of expected query parameters and their expected types. 
//Returns slice of QueryValidationError if validation fails
func (qv *Service) ValidateQuery(queries url.Values, rules map[string]string) (map[string]string, []QueryValidationError) {
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

		ok, err := qv.ValidateParamValue(value, expectedType)
		if err != nil {
			errors = append(errors, QueryValidationError{
				Parameter: param,
				Value: value,
				Message: err.Error(),
			})
		}
		if !ok {
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
func (qv *Service) BuildSqlQuery(queries map[string]string, accountID string) (string, []any, error) {
	query := "SELECT * FROM transactions WHERE account_id = $1"
	args := []any{accountID}
	paramCount := 2

	if val, ok := queries["merchant"]; ok {
		if val != "" {
			query += fmt.Sprintf(" AND merchant_name ILIKE $%d", paramCount)
			args = append(args, "%"+val+"%")
			paramCount++
		}
	}
	if val, ok := queries["category"]; ok {
		if val != "" {
			query += fmt.Sprintf(" AND personal_finance_category ILIKE $%d", paramCount)
			args = append(args, "%"+val+"%")
			paramCount++
		}
	}
	if val, ok := queries["channel"]; ok {
		if val != "" {
			query += fmt.Sprintf(" AND payment_channel ILIKE $%d", paramCount)
			args = append(args, "%"+val+"%")
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

	query += " ORDER BY date DESC"
	
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