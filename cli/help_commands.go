package main

import (
	"log"
	"github.com/jms-guy/greed/cli/internal/config"
)

//Function will execute different help functions depending on argument
func commandHelp(c *config.Config, args []string) error {
	//Base help function
	if len(args) == 0 {
		helpBase()
		return nil
	}

	//Too many arguments
	if len(args) > 1 {
		log.Println("Too many arguments given - type --help for more details")
		return nil
	}

	cmdArg := args[0]

	//Users help func
	if cmdArg == "users" {
		helpUsers()
		return nil
	}

	//Accounts help func
	if cmdArg == "accounts" {
		helpAccounts()
		return nil
	}

	//Transactions help func
	if cmdArg == "transactions" {
		helpTransactions()
		return nil
	}

	if cmdArg == "all" {
		helpAll()
		return nil
	}

	return nil
}

//Function will list more specific help commands
func helpBase() {
	log.Println("\r                                              ")
	log.Println("\r~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	log.Println("\r  Greed is a financial tracking and planning application. If you are a new user, create a new user first.")
	log.Println("\r Once you have created a user, you may create accounts to attach to the user, to be used to track ")
	log.Println("\r financial data. For specific commands, see below. ")
	log.Println("\r Do not enter any (), [], or {} characters in commands.")
	log.Println("\r~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	log.Println("\r  - For user commands, enter 		[help users]")
	log.Println("\r  - For account commands, enter 	[help accounts]")
	log.Println("\r  - For transaction commands, enter 	[help transactions]")
	log.Println("\r  - For a full list of commands, enter	[help all]")
	log.Println("\r                                              ")
}

func helpUsers() {
	log.Println("\r                                              ")
	log.Println("\r~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	log.Println("\r  Username must be unique. Password input must be greater than 8 characters.")
	log.Println("\r  Below is a list of user commands.")
	log.Println("\r  Do not enter any (), [], or {} characters in commands.")
	log.Println("\r~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	log.Println("\r  - To create a new user, enter			[user-create {name}]")
	log.Println("\r  - To delete a user, enter			[user-delete {name}]")
	log.Println("\r  - To obtain a list of local users, enter	[users]")
	log.Println("\r  - To login as a user, enter			[user-login {name}]")
	log.Println("\r  - To logout of a user, enter			[user-logout]")
	log.Println("\r                                              ")
}

func helpAccounts() {
	log.Println("\r                                              ")
	log.Println("\r~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	log.Println("\r  Multiple accounts allowed per user. Account names must be unique per user.")
	log.Println("\r  Below is a list of account commands.")
	log.Println("\r  Do not enter any (), [], or {} characters in commands.")
	log.Println("\r~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	log.Println("\r  - To create an account, enter					[account-create {name}]")
	log.Println("\r  - To delete an account, enter					[account-delete {name}]")
	log.Println("\r  - To obtain a list of account attached to user, enter 	[accounts]")
	log.Println("\r  - To get the name of the currently logged in account, enter	[account]")
	log.Println("\r  - To login to an account, enter				[account-login {name}]")
	log.Println("\r  - To logout of an account, enter				[account-logout]")
	log.Println("\r                                              ")
}

func helpTransactions() {
	
}

func helpAll() {
	log.Println("\r                                              ")
	log.Println("\r  - To create a new user, enter					[user-create {name}]")
	log.Println("\r  - To delete a user, enter					[user-delete {name}]")
	log.Println("\r  - To obtain a list of local users, enter			[users]")
	log.Println("\r  - To login as a user, enter					[user-login {name}]")
	log.Println("\r  - To logout of a user, enter					[user-logout]")
	log.Println("\r  - To create an account, enter					[account-create {name}]")
	log.Println("\r  - To delete an account, enter					[account-delete {name}]")
	log.Println("\r  - To obtain a list of account attached to user, enter 	[accounts]")
	log.Println("\r  - To get the name of the currently logged in account, enter	[account]")
	log.Println("\r  - To login to an account, enter				[account-login {name}]")
	log.Println("\r  - To logout of an account, enter				[account-logout]")
	log.Println("\r                                              ")
}

/*
func showCurrencies() {
	log.Println("\rCurrently supported currencies: ")
	log.Println("\r [CAD] [USD] [EUR] [GBP]        ")
}

//A currency enum struct, containing the valid currencies supported in this application
type Currency int

const (
	CAD Currency = iota
	USD
	EUR
	GBP
)

//Mapping currency strings to enum 
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
}*/