package main

import (
	"strings"
)

//Cleans user input string for command arguments use
func cleanInput(s string) []string {	
	lowerS := strings.ToLower(s)
	results := strings.Fields(lowerS)
	return results
}