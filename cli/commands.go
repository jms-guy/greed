package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jms-guy/greed/models"
)

func commandCreateUser(c *Config, args []string) (error) {
	//Make sure enough arguments present
	if len(args) < 1 {
		log.Println("Missing argument - type {syntax} for command structure")
		return nil
	}

	//Username is first argument
	username := args[0]
	url := c.Client.baseURL + "/api/users"

	//Create a bufio scanner for getting password
	var password string
	scanner := bufio.NewScanner(os.Stdin)

	//Starts a scanner loop for password input
	for {
		fmt.Print("Please enter a password > ")
		scanner.Scan()

		pw := scanner.Text()
		if len(pw) < 8 {
			fmt.Println("Password must be greater than 8 characters")
			continue
		} else {
			password = pw
			break
		}
	}

	//Create request struct
	reqData := models.UserDetails{
		Name: username,
		Password: password,
	}

	//Marshal data
	dataToSend, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("error marshalling json data: %w", err)
	}

	//Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dataToSend))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	//Send request
	res, err := c.Client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	//Check response status code
	if res.StatusCode >= 400 {
		return fmt.Errorf("server returned error: %s", res.Status)
	}

	//Decode response data
	var user models.User
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}
	/*
	//Create a quick slice to pack user.Name into to pass into commandUserLogin
	var s []string
	s[0] = user.Name
	err = commandUserLogin(c, s)
	if err != nil {
		return fmt.Errorf("error logging into user %s: %w", user.Name, err)
	}
*/
	return nil
}
/*
func commandUserLogin(c *Config, args []string) error {
	//Case for being called after the creation of a new user, in which case the new user's name
	//field is directly passed as the argument
	if len(args) == 1 {
		if args[0] != "user-login" {
			c.CurrentUserName = args[0]
			return nil
		}
	}
	
}
*/