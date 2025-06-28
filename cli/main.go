package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/jms-guy/greed/cli/internal/config"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

//Don't forget, update bufio scanners to handle sigInt

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
	}

	args := os.Args

	if len(args) < 3 {
		fmt.Println("Command Usage: greed [registry] <subcommand> {arguments}")
		fmt.Println("Example Command: greed user register James")
		fmt.Println("Type {greed --help} for more details")
		os.Exit(1)
	}

	//args[0] = program name, args[1] = registry name, args[2] = command name, args[3:] = arguments/flags
 	registry := args[1]
	command := args[2]

	if registry == "user" {
		cmd, ok := userRegistry[command]
		if !ok {
			log.Fatalf("Command not found")
		}
		err = cmd.callback(cfg, args[3:])
		if err != nil {
			log.Fatalf("%s\n", err)
		}
	}  else if  registry == "item" {
		cmd, ok := itemRegistry[command]
		if !ok {
			log.Fatalf("Command not found")
		}
		err = cmd.callback(cfg, args[3:])
		if err != nil {
			log.Fatalf("%s\n", err)
		}
	} else if registry == "admin" {
		cmd, ok := adminRegistry[command]
		if !ok {
			log.Fatalf("Command not found")
		}
		err = cmd.callback(cfg, args[3:])
		if err != nil {
			log.Fatalf("%s\n", err)
		}
	}

}