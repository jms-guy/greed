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

var userRegistry map[string]cliCommand

func init() {
	userRegistry = map[string]cliCommand{
		"create":	{
			name: "create",
			description: "Create a new user",
			callback: commandCreateUser,
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
}