package main

import (
	
)

//A resgistry struct, that holds all of the commands that the CLI supports, as
//well as their descriptions, and callback functions associated with them.

type cliCommand struct {
	name			string
	description		string
	syntax			string
	callback		func(c *Config, args []string) (error)
}

var commandRegistry map[string]cliCommand

func init() {
	commandRegistry = map[string]cliCommand{
		"create-user":	{
			name: "create-user",
			description: "Creates a user in the database",
			syntax: "{create-user} {name}",	
			callback: commandCreateUser,
		},
		"delete-user":	{
			name: "delete-user",
			description: "Deletes a user from the database",
			syntax: "{delete-user} {name}",
			callback: commandDeleteUser,
		},
		"create-account":	{
			name: "create-account",
			description: "Creates an account for current user",
			syntax: "{create-account} {name}",
			callback: commandCreateAccount,
		},
		"delete-account":	{
			name: "delete-account",
			description: "Deletes an account attached to user",
			syntax: "{delete-account} {name}",
			callback: commandDeleteAccount,
		},
		"user-login":	{
			name: "login",
			description: "Login as specified user",
			syntax: "{user-login} {name}",
			callback: commandUserLogin,
		},
		"user-logout":	{
			name: "logout",
			description: "Logs out of user credentials",
			syntax: "{user-logout} {name}",
			callback: commandUserLogout,
		},
		"account-login":	{
			name: "account-login",
			description: "Logs into a user's specified account",
			syntax: "{account-login} {name}",
			callback: commandAccountLogin,
		},
		"account-logout":	{
			name: "account-logout",
			description: "Logs out of a user's specified account",
			syntax: "{account-logout} {name}",
			callback: commandAccountLogout,
		},
		"exit":	{
			name: "exit",
			description: "Exits the program",
			syntax: "{exit}",
			callback: commandExit,
		},
		"help": {
			name: "help",
			description: "Prints list of commands and their descriptions",
			syntax: "{help}",
			callback: commandHelp,
		},
		"syntax": {
			name: "syntax",
			description: "Provides a list of commands and their required syntax",
			syntax: "{syntax}",
			callback: commandSyntax,
		},
		"reset":	{
			name: "reset",
			description: "Admin testing tool, resets database tables",
			syntax: "{reset} {table}",
			callback: commandReset,
		},
	}
}