package main

import (
	"log"
	"os"
	"github.com/joho/godotenv"
	"github.com/jms-guy/greed/cli/internal/config"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
	}	

	//Get user input arguments
	args := os.Args
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
	} 

}