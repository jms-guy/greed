package main

import (
	"github.com/jms-guy/greed/cli/internal/config"
)

//A resgistry struct, that holds all of the commands that the CLI supports, as
//well as their descriptions, and callback functions associated with them.

type cliCommand struct {
	name			string
	description		string
	callback		func(c *config.Config, args []string) error
}

var commandRegistry map[string]cliCommand

func init() {
	commandRegistry = map[string]cliCommand{
		"user-create":	{
			name: "user-create",
			description: "Creates a user in the database",
			callback: commandCreateUser,
		},
		"user-login":	{
			name: "login",
			description: "Login as specified user",
			callback: commandUserLogin,
		},
		"user-logout":	{
			name: "logout",
			description: "Logs out of user credentials",
			callback: commandUserLogout,
		},
		"user-delete":	{
			name: "user-delete",
			description: "Deletes a user from the database",
			callback: commandDeleteUser,
		},
		"users":	{
			name: "users",
			description: "Gets a list of local users",
			callback: commandGetUsers,
			},
		"help": {
			name: "help",
			description: "Prints list of commands and their descriptions",
			callback: commandHelp,
		},
		"account-create":	{
			name: "account-create",
			description: "Creates an account for current user",
			callback: commandCreateAccount,
		},
		"account-login":	{
			name: "account-login",
			description: "Logs into a user's specified account",
			callback: commandAccountLogin,
		},
		"account-logout":	{
			name: "account-logout",
			description: "Logs out of a user's specified account",
			callback: commandAccountLogout,
		},
		"account": {
			name: "account",
			description: "Shows the name of the currently logged in account",
			callback: commandGetCurrentAccount,
		},
		"accounts": {
			name: "accounts",
			description: "Shows the names of accounts attached to current user",
			callback: commandListAccounts,
		},
		"account-delete":	{
			name: "account-delete",
			description: "Deletes an account attached to user",
			callback: commandDeleteAccount,
		},
	}
}