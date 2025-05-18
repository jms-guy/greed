package main

import (
	"bufio"
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
		CurrentUserName: "",
		Client: *NewClient(addr),
	}


	//Create stdin text scanner
	scanner := bufio.NewScanner(os.Stdin)

	//Start up the client
	log.Println("Starting Client...")
	for {
		log.Printf("Greed > ")

		//Wait for text input, when received, clean the input
		scanner.Scan()
		userInput := cleanInput(scanner.Text())
		if len(userInput) == 0 {
			continue
		}

		//Search command registry for command given
		command, ok := commandRegistry[userInput[0]]
		if !ok {
			log.Println("Unknown command")
			continue
		}

		//Execute command callback
		err := command.callback(&cfg, userInput)
		if err != nil {
			log.Printf("Returned error: %s", err)
			continue
		}
	}
}