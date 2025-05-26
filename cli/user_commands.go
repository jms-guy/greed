package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/config"
	"github.com/jms-guy/greed/models"
)

//Creates a user record in the database, as well as a config file for that user
func commandCreateUser(c *config.Config, args []string) error {
	if len(args) < 1 {
		log.Println("\rMissing argument - type help for details")
		return nil
	}

	username := args[0]

	users, err := c.GetUsers()
	if err != nil {
		return err
	}
	if slices.Contains(users, username) {
		log.Println("\rUser already exists on this machine")
		return nil
	}

	//Set api endpoint 
	url := c.Client.BaseURL + "/api/users"

	//Create a bufio scanner for getting password
	var password string

	for {
		pw, err := auth.ReadPassword("Please enter a password > ")
		if err != nil {
			return fmt.Errorf("error getting password: %w", err)
		}

		if len(pw) < 8 {
			fmt.Println(" ")
			fmt.Println("Password must be greater than 8 characters")
			fmt.Println("")
			continue
		} else {
			//Confirm password input
			for {
				confirmPw, err := auth.ReadPassword("Confirm password > ")
				if err != nil {
					return fmt.Errorf("error reading confirmed password: %w", err)
				}

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

	if res.StatusCode >= 400 {
		return fmt.Errorf("server returned error: %s", res.Status)
	}

	var user models.User
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	//Create user config file in .config directory
	err = c.CreateUser(user)
	if err != nil {
		return err
	}

	err = c.CreateAccountsConfigFilePath(username)
	if err != nil {
		return err
	}
	
	log.Println("\rUser created successfully!")
	log.Println("\r                                          ")

	return nil
}

//Function will take a username and "log in", moving that user's config file to the current session directory
//where it will be loaded into the Config struct on each command
func commandUserLogin(c *config.Config, args []string) error {
	//Set url for request endpoint
	loginUrl := c.Client.BaseURL + "/api/users"

	if len(args) > 2 {
		return fmt.Errorf("too many arguments - type help for details")
	}

	//If logged in as another user, can't log in
	if c.FileData.User.Name != "" {
		fmt.Printf("\rAlready logged in as user: %s. Please log out first\n", c.FileData.User.Name)
		return nil
	}

	username := args[0]

	users, err := c.GetUsers()
	if err != nil {
		return err
	}
	if !slices.Contains(users, username) {
		log.Print("\rUser does not exist on this machine")
		return nil
	}

	//Create empty password variable to get input, and empty user struct to decode json response into
	//maxretries and attempts for getting password input, as well as reading config file
	var user models.User
	password := ""
	maxRetries := 3
	attempts := 0	

	for attempts < maxRetries{
		password, err =auth.ReadPassword("Please enter password > ")
		if err != nil {
			return fmt.Errorf("error reading password: %w", err)
		}

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
	
		if res.StatusCode == 400 {
			attempts++
			fmt.Println("Incorrect password")
			continue
		}

		if res.StatusCode >= 500 {
			return fmt.Errorf("server returned error: %s", res.Status)
		}
	
		if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
			return fmt.Errorf("error decoding response data: %w", err)
		}

		break
	}

	if attempts >= maxRetries {
		return fmt.Errorf("max password attempts reached")
	}

	//Set current session file to user
	err = c.SetCurrentUserSession(username)
	if err != nil {
		return fmt.Errorf("error setting current session: %w", err)
	}

	log.Printf("\rLogged in successfully as: %s\n", username)
	return nil
}

//Function gets a list of local users
func commandGetUsers(c *config.Config, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("command takes no arguments - type help for details")
	}

	users, err := c.GetUsers()
	if err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Println("\rNo user files currently on this machine.")
		return nil
	}

	fmt.Println("\rUsers on this machine: ")
	for _, user := range users {
		fmt.Printf(" > %s \n", user)
	}

	return nil
}

//Logs out a user from the current session, moving their config file back to the /users config directory
func commandUserLogout(c *config.Config, args []string) error {
	if len(args) >= 1 {
		return fmt.Errorf("too many arguments provided - type help for details")
	}

	if c.FileData.User.Name == "" {
		return fmt.Errorf("no user is currently logged in")
	}

	if err := c.EndCurrentUserSession(); err != nil {
		return fmt.Errorf("error ending current session: %w", err)
	}

	log.Printf("\r%s logged out successfully\n", c.FileData.User.Name)

	c.FileData = config.FileData{}

	return nil
}


//Function will delete a user record in the database, along with their config file
func commandDeleteUser(c *config.Config, args []string) error {
	//Set api endpoint url
	getUrl := c.Client.BaseURL + "/api/users"

	//Check arguments length
	if (len(args) == 0) || (len(args) > 1) {
		return fmt.Errorf("incorrect number of arguments given - type help for details")
	}

	//Get username
	username := args[0]

	//Check username against current session data
	if c.FileData.User.Name == "" {
		log.Println("\rMust be logged in to use this command")
		return nil
	}
	if username != c.FileData.User.Name {
		log.Printf("\rCannot delete user: %s\n", username)
		log.Println("\rMust be logged in to delete user records")
		log.Printf("\rCurrently logged in as user: %s\n", c.FileData.User.Name)
		return nil
	}

	password := ""
	var err error
	maxRetries := 3
	attempts := 0

	//Starts a scanner loop for password input
	for attempts < maxRetries{
		password, err = auth.ReadPassword("Please enter password > ")
		if err != nil {
			return fmt.Errorf("error getting password: %w", err)
		}

		reqData := models.UserDetails{
			Name: username,
			Password: password,
		}

		//Make request
		res, err := c.MakeBasicRequest("GET", getUrl, reqData)
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

		break
	}

	//If incorrect password maxtries was reached
	if attempts >= maxRetries {
		return fmt.Errorf("max password attempts reached")
	}

	scanner := bufio.NewScanner(os.Stdin)
	//Start another scanner loop for confirmation input
	for {
		fmt.Printf("Are you sure you want to delete user: %s? [y, n]\n", username)
		fmt.Print(" > ")
		scanner.Scan()

		if scanner.Text() == "n" {
			fmt.Println("\rUser deletion aborted")
			return nil
		} else if scanner.Text() != "y" {
			fmt.Println("\rPlease enter either [y] or [n]")
			continue
		} else {
			fmt.Printf("\rCommencing deletion of user: %s\n", username)
			break
		}
	}

	//Set delete api endpoint url
	deleteUrl := c.Client.BaseURL + "/api/users/" + c.FileData.User.ID.String()

	//Delete local user config file
	fmt.Println("\rDeleting user's config file...")
	err = c.DeleteUser(username)
	if err != nil {
		return err
	}
	fmt.Println("\rUser's config file deleted successfully!")

	//Make call to server database to delete records
	fmt.Println("\rDeleting user's database records...")
	response, err := c.MakeBasicRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return err 
	}
	defer response.Body.Close()

	if response.StatusCode > 400 {
		return fmt.Errorf("server returned error: %s", response.Status)
	}

	fmt.Println("\rUser has been deleted successfully!")

	return nil
}