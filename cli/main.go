package main

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

func main() {
	//Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	//.env Website URL
	addr := os.Getenv("ADDRESS")

	//Create config struct
	cfg := Config{
		FileData: FileData{},
		EnvData: EnvData{
			Address: addr,
		},
		Client: *NewClient(addr),
	}

	//Get user input arguments
	args := os.Args
	//args[0] = program name, args[1] = command name, args[2:] = arguments
 	command := args[1]

	//Check if input command is in command registry
	cmd, ok := commandRegistry[command]
	if !ok {
		log.Println("Command not found")
	}

	//Execute command callback function
	err = cmd.callback(&cfg, args[2:])
	if err != nil {
		log.Fatalf("\r%s\n", err)
	}
}