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

var userRegistry 	map[string]cliCommand
var adminRegistry 	map[string]cliCommand

func init() {
	userRegistry = map[string]cliCommand{
		"register":	{
			name: "register",
			description: "Register a new user",
			callback: commandRegisterUser,
		},
		"login":	{
			name: "login",
			description: "Login as specified user",
			callback: commandUserLogin,
		},
		"logout":	{
			name: "logout",
			description: "Logs out of user credentials",
			callback: commandUserLogout,
		},
		"delete":	{
			name: "user-delete",
			description: "Deletes a user from the database",
			callback: commandDeleteUser,
		},
	}

	adminRegistry = map[string]cliCommand{
		"clear": {
			name: "clear",
			description: "Clear local database",
			callback: commandAdminClear,
		},
	}
}