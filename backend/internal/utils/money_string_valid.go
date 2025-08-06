package utils

import (
	"regexp"
)

// Validation function that uses regex to make sure that a given string is in the format
// 'xxx.xx' to prevent bad inputs into the database
func moneyStringValidation(s string) bool {
	matched, _ := regexp.MatchString(`^(\d+)(\.\d{2})?$`, s)
	return matched
}
