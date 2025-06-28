package main

import (
	"fmt"
	"log"

	"github.com/jms-guy/greed/cli/internal/config"
)

//A resgistry struct, that holds all of the commands that the CLI supports, as
//well as their descriptions, and callback functions associated with them.

type cliCommand struct {
	name			string
	description		string
	usage			string
	callback		func(c *config.Config, args []string) error
}

var userRegistry 			map[string]cliCommand
var accountsRegistry 		map[string]cliCommand
var itemRegistry   			map[string]cliCommand
var transactionsRegistry 	map[string]cliCommand
var adminRegistry 			map[string]cliCommand

func FindRegistry(cfg *config.Config, registry, command string, args []string) error {
	var err error
	switch registry {
	case "user":
		cmd, ok := userRegistry[command]
		if !ok {
			log.Fatalf("Command not found")
		}
		if err = cmd.callback(cfg, args); err != nil {
			return err 
		}
		return nil
	case "item":
		cmd, ok := itemRegistry[command]
		if !ok {
			log.Fatalf("Command not found")
		}
		if err = cmd.callback(cfg, args); err != nil {
			return err 
		}
		return nil
	case "accounts":
		cmd, ok := accountsRegistry[command]
		if !ok {
			log.Fatalf("Command not found")
		}
		if err = cmd.callback(cfg, args); err != nil {
			return err 
		}
		return nil
	case "admin":
		cmd, ok := adminRegistry[command]
		if !ok {
			log.Fatalf("Command not found")
		}
		if err = cmd.callback(cfg, args); err != nil {
			return err 
		}
		return nil
	case "transactions":
		cmd, ok := transactionsRegistry[command]
		if !ok {
			log.Fatalf("Command not found")
		}
		if err = cmd.callback(cfg, args); err != nil {
			return err 
		}
		return nil
	default:
		fmt.Println("Registry not found")
		return nil
	}
}

func init() {
	userRegistry = map[string]cliCommand{
		"register":	{
			name: "register",
			description: "Register a new user",
			usage: "greed user register {name}",
			callback: commandRegisterUser,
		},
		"login":	{
			name: "login",
			description: "Login as specified user",
			usage: "greed user login {name}",
			callback: commandUserLogin,
		},
		"logout":	{
			name: "logout",
			description: "Logs out of user credentials",
			usage: "greed user logout",
			callback: commandUserLogout,
		},
		"delete":	{
			name: "delete",
			description: "Deletes a user from the database",
			usage: "greed user delete {name}",
			callback: commandDeleteUser,
		},
		"verify":	{
			name: "verify",
			description: "Verifies the user's email address",
			usage: "greed user verify",
			callback: commandVerifyEmail,
		},
		"items": 	{
			name: "items",
			description: "Lists a user's item records",
			usage: "greed user items",
			callback: commandUserItems,
		},
		"updatepw": {
			name: "updatepw",
			description: "Updates a user's password",
			usage: "greed user updatepw",
			callback: commandUpdatePassword,
		},
		"resetpw": 	{
			name: "resetpw",
			description: "Resets a user's forgotten password",
			usage: "greed user resetpw {email}",
			callback: commandResetPassword,
		},
	}

	itemRegistry = map[string]cliCommand{
		"accounts": {
			name: "accounts",
			description: "Get all accounts for an item",
			usage: "greed item accounts {item name}",
			callback: commandGetAccounts,
		},
		"transactions": {
			name: "transactions",
			description: "Retrieves transaction records for all accounts for item on first fetch. Subsequent calls sync new transaction data",
			usage: "greed item transactions {item name}",
			callback: commandGetTransactions,
		},
		"rename": {
			name: "rename",
			description: "Rename an item",
			usage: "greed item rename {current name} {new name}",
			callback: commandRenameItem,
		},
		"delete": {
			name: "delete",
			description: "Delete an item record",
			usage: "greed item delete {name}",
			callback: commandDeleteItem,
		},
	}

	accountsRegistry = map[string]cliCommand{
		"list": {
			name: "list",
			description: "List all accounts for item",
			usage: "greed accounts list {item name}",
			callback: commandListAccounts,
		},
		"listall": {
			name: "listall",
			description: "Lists all accounts for user",
			usage: "greed accounts listall",
			callback: commandListAllAccounts,
		},
		"info": {
			name: "info",
			description: "Lists extended account information for given account name",
			usage: "greed accounts info {account name}",
			callback: commandAccountInfo,
		},
	}

	transactionsRegistry = map[string]cliCommand{
		"get": {
			name: "get",
			description: "Gets list of transactions for account. Takes many optional arguments, type [--help transactions] for more details",
			usage: "greed transactions get {account name}",
			callback: commandGetTxnsAccount,
		},
	}

	adminRegistry = map[string]cliCommand{
		"clear": {
			name: "clear",
			description: "Clear local database",
			usage: "greed admin clear",
			callback: commandAdminClear,
		},
	}
}