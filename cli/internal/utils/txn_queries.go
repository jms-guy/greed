package utils

import (
	"fmt"
	"strings"
)

//Function maps out query flags to values, assuming a structure of "-flag value -flag value -flag value"
//ex. "--merchant uber --category travel --limit 10"
func BuildQueries(args []string) (string, error) {
	queries := make(map[string]string)

	for i := 0; i < len(args) - 1; i += 2 {
		queries[args[i]] = args[i + 1]
	}
	if len(args)%2 != 0 {
		return "", fmt.Errorf("odd number of arguments for flag/value pairs")
	}

	for key, val := range queries {
		if !strings.HasPrefix(key, "--") {
			return "", fmt.Errorf("improper query argument syntax - type '--help transactions' for more details")
		}
		if strings.HasPrefix(val, "--") {
			return "", fmt.Errorf("improper query argument syntax - type '--help transactions' for more details")
		}
	}

	queryString := "?"
	for key, value := range queries {
		queryString += fmt.Sprintf("%s=%s&", key[2:], value)
	}

	queryString = strings.TrimRight(queryString, "&")

	return queryString, nil
}