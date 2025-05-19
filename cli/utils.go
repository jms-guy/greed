package main

import (
	"strings"
	"fmt"
	"bufio"
	"os"
)

//Cleans user input string for command arguments use
func cleanInput(s string) []string {	
	lowerS := strings.ToLower(s)
	results := strings.Fields(lowerS)
	return results
}

func getPassword() string {
	var password string
	scanner := bufio.NewScanner(os.Stdin)

	//Starts a scanner loop for password input
	for {
		fmt.Print("Please enter password > ")
		scanner.Scan()

		pw := scanner.Text()
		if len(pw) < 8 {
			fmt.Println("Password must be greater than 8 characters")
			continue
		} else {
			//Confirm password input
			for {
				fmt.Print("Confirm password > ")
				scanner.Scan()
				confirmPw := scanner.Text()

				if confirmPw == pw {
					password = pw
					break
				} else {
					fmt.Println("Password does not match")
					continue
				}
			}
			break
		}
	}

	return password
}