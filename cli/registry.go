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
var accountsRegistry map[string]cliCommand
var itemRegistry   map[string]cliCommand
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
			name: "delete",
			description: "Deletes a user from the database",
			callback: commandDeleteUser,
		},
		"verify":	{
			name: "verify",
			description: "Verifies the user's email address",
			callback: commandVerifyEmail,
		},
		"items": 	{
			name: "items",
			description: "Lists a user's item records",
			callback: commandUserItems,
		},
		"updatepw": {
			name: "updatepw",
			description: "Updates a user's password",
			callback: commandUpdatePassword,
		},
		"resetpw": 	{
			name: "resetpw",
			description: "Resets a user's forgotten password",
			callback: commandResetPassword,
		},
	}

	itemRegistry = map[string]cliCommand{
		"rename": {
			name: "rename",
			description: "Rename an item",
			callback: commandRenameItem,
		},
		"delete": {
			name: "delete",
			description: "Delete an item record",
			callback: commandDeleteItem,
		},
	}

	accountsRegistry = map[string]cliCommand{
		"get": {
			name: "get",
			description: "Get all accounts for an item",
			callback: commandGetAccounts,
		},
		"list": {
			name: "list",
			description: "List all accounts for item",
			callback: commandListAccounts,
		},
		"listall": {
			name: "listall",
			description: "Lists all accounts for user",
			callback: commandListAllAccounts,
		},
		"info": {
			name: "info",
			description: "Lists account information for given account name",
			callback: commandAccountInfo,
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