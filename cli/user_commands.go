package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/models"
)

func commandCreateUser(c *Config, args []string) (error) {
	//Make sure enough arguments present
	if len(args) < 1 {
		log.Println("\rMissing argument - type help for details")
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
			//Confirm password input
			for {
				fmt.Print("Confirm password > ")
				scanner.Scan()
				confirmPw := scanner.Text()

				if confirmPw == pw {
					password = pw
					break
				} else {
					fmt.Println("Password does not match")
					continue
				}
			}
			break
		}
	}

	//Create request struct
	reqData := models.UserDetails{
		Name: username,
		Password: password,
	}

	//Make request
	res, err := c.MakeBasicRequest("POST", url, reqData)
	if err != nil {
		return err
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
	
	log.Println("\rUser created successfully!")
	return nil
}

func commandUserLogin(c *Config, args []string) error {
	//Create empty username variable, and set url for request endpoint
	loginUrl := c.Client.baseURL + "/api/users"
	searchUsersUrl := c.Client.baseURL + "/api/users/all"

	//Too many arguments provided
	if len(args) > 2 {
		return fmt.Errorf("too many arguments - type help for details")
	}

	//Set username
	username := args[0]

	//Check database for list of users, make sure username is present in list
	response, fErr := c.MakeBasicRequest("GET", searchUsersUrl, nil)
	if fErr != nil {
		return fErr
	}
	defer response.Body.Close()

	if response.StatusCode > 400 {
		return fmt.Errorf("server returned error: %s", response.Status)
	}

	var users []string
	if err := json.NewDecoder(response.Body).Decode(&users); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	if !slices.Contains(users, username) {
		log.Print("\rUser does not exist in database")
		return nil
	}

	//Create empty password variable to get input, and empty user struct to decode json response into
	//maxretries and attempts for getting password input, as well as reading config file
	var user models.User
	password := ""
	maxRetries := 3
	attempts := 0

	//New bufio scanner for password input
	scanner := bufio.NewScanner(os.Stdin)	

	//Starts a scanner loop for password input
	for attempts < maxRetries{
		fmt.Print("Please enter password > ")
		scanner.Scan()

		password = scanner.Text()

		reqData := models.UserDetails{
			Name: username,
			Password: password,
		}

		//Make request
		res, err := c.MakeBasicRequest("GET", loginUrl, reqData)
		if err != nil {
			return err
		}
		defer res.Body.Close()
	
		//If incorrect password was entered, ask for password again
		if res.StatusCode == 400 {
			attempts++
			fmt.Println("Incorrect password")
			continue
		}

		//Check response status code for server errors
		if res.StatusCode >= 500 {
			return fmt.Errorf("server returned error: %s", res.Status)
		}
	
		//Decode response data
		if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
			return fmt.Errorf("error decoding response data: %w", err)
		}

		break
	}

	//If incorrect password maxtries was reached
	if attempts >= maxRetries {
		return fmt.Errorf("max password attempts reached")
	}

	//Create variables for use in reading .config file
	var fileData FileData
	var err error

	//Read config file for data
	fileData, err = Read(username)
	if err != nil || fileData.User.ID == uuid.Nil {
		err = c.SetUser(user)
		if err != nil {
			return err
		}

		fileData, err = Read(username)
		if err != nil {
			return err
		}
	}

	//Return config file data
	c.FileData = fileData

	log.Printf("\rLogged in successfully as: %s\n", username)
	return nil
}

func commandUserLogout(c *Config, args []string) error {
	if len(args) > 2 {
		return fmt.Errorf("too many arguments provided - type help for details")
	}

	username := args[0]

	if username == c.FileData.User.Name {
		c.FileData = FileData{}
	} else {
		return fmt.Errorf("incorrect username")
	}

	return nil
}
