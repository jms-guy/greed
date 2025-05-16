package main

import "fmt"

//A currency enum struct, containing the valid currencies supported in this application
type Currency int

const (
	CAD Currency = iota
	USD
	EUR
	GBP
)

var currencyToString = map[Currency]string{
	CAD:	"CAD",
	USD:	"USD",
	EUR:	"EUR",
	GBP:	"GBP",
}

var stringToCurrency = map[string]Currency{
	"CAD":	CAD,
	"USD":	USD,
	"EUR":	EUR,
	"GBP":	GBP,
}

//Function for validating a given currency string
func isValidCurrencyString(s string) bool {
	_, ok := stringToCurrency[s]
	return ok
}


func stringToCurrencyValue(s string) (Currency, error) {
	if c, ok := stringToCurrency[s]; ok {
		return c, nil
	}
	return 0, fmt.Errorf("invalid currency: %s", s)
}