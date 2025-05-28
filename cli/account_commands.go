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

//Function will create a new account for user
func commandCreateAccount(c *config.Config, args []string) error {
	if c.FileData.Account.Name != "" {
		log.Println("\rCurrently logged into an account, log out to create a new one.")
		return nil
	}

	if len(args) < 1 {
		log.Println("\rMissing argument - type help for details")
		return nil
	}

	accountName := args[0]

	accounts, err := c.GetAccounts()
	if err != nil {
		return err
	}

	if slices.Contains(accounts, accountName) {
		log.Printf("Account named '%s' already exists for user", accountName)
		return nil
	}

	if c.FileData.Account.Name == accountName {
		log.Printf("Account named '%s' already exists for user", accountName)
		return nil
	}

	userId := c.FileData.User.ID

	//Set api endpoint url
	url := c.Client.BaseURL + "/api/users/" + userId.String() + "/accounts"

	//Changed my mind on having manual account's right off the bat, maybe re-add them later
	/*
	log.Println("\r~Creating account:               ")
	log.Println("\rWill this account be synced to a financial institution, or will the data be manually input? ")
	log.Printf("\r     [auto]/[manual]     ")

	//Start bufio scanner for additional account fields(InputType and Currency)
	inputType := ""
	currencyType := ""

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\r                    > ")
		scanner.Scan()

		input := scanner.Text()
		if (input != "auto") && (input != "manual") {
			log.Printf("\rPlease enter either [auto] or [manual]\n")
			continue
		}

		if input == "auto" {
			inputType = "auto"
		} else if input == "manual" {
			inputType = "manual"
		}

		break
	}
		

	scanner := bufio.NewScanner(os.Stdin)
	currencyType := ""

	log.Println("\r                             ")
	log.Println("\rPlease set the base currency for this account. May be changed at later.")
	log.Println("\rIf no currency is entered, it will be defaulted to 'CAD'. For a list of supported currencies, type: [currencies]")

	for {
		fmt.Print("\r                   > ")
		scanner.Scan()

		currency := scanner.Text()
		if currency == "currencies" {
			showCurrencies()
			continue
		}

		if isValidCurrencyString(currency) {
			currencyType = currency
		} else {
			log.Println("\rUnsupported currency entered")
			continue
		}

		break
	}*/

	//Create account request
	reqParams := models.AccountDetails{
		Name: accountName,
		UserID: userId,
	}

	res, err := c.MakeBasicRequest("POST", url, reqParams)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 400 {
		return fmt.Errorf("error in response data: %s", res.Status)
	}

	var account models.Account
	if err := json.NewDecoder(res.Body).Decode(&account); err != nil {
		return err
	}

	c.FileData.Account = account

	//Save updated FileData to the user config file in the currentsession directory
	err = c.CreateAccount()
	if err != nil {
		return fmt.Errorf("error saving file data: %w", err)
	}

	log.Println("\rAccount created successfully!")
	log.Println("\r                                            ")

	return nil
}

//Logs into a user's specified account
func commandAccountLogin(c *config.Config, args []string) error {
	if len(args) != 1 {
		log.Println("\rCommand syntax issue - type help for more details")
		return nil
	}

	if c.FileData.Account.Name != "" {
		log.Println("\rAlready logged into an account, cannot log into a second one")
		return nil
	}

	accountName := args[0]

	accounts, err := c.GetAccounts()
	if err != nil {
		return err
	}

	if !slices.Contains(accounts, accountName) {
		log.Println("\rNo account found with that name")
		return nil 
	}

	if err = c.MoveAccountToCurrentSession(accountName); err != nil {
		return err 
	}

	log.Printf("\rSuccessfully logged into account: %s\n", accountName)
	return nil
}

//Logs out of a user's account, leaving just the user logged in
func commandAccountLogout(c *config.Config, args []string) error {
	if len(args) >= 1 {
		return fmt.Errorf("too many arguments provided - type help for details")
	}

	if c.FileData.Account.Name == "" {
		log.Println("\rNot currently logged into an account.")
		return nil
	}

	err := c.MoveAccountToUsersDir()
	if err != nil {
		return fmt.Errorf("error logging out of account: %w", err)
	}

	log.Println("\rSuccessfully logged out of account!")

	return nil
}

//Returns a list of accounts attached to logged in user
func commandListAccounts(c *config.Config, args []string) error {
	if len(args) != 0 {
		log.Println("Command syntax issue - type help for more details")
		return nil 
	}

	accounts, err := c.GetAccounts()
	if err != nil {
		return fmt.Errorf("error getting user's accounts: %w", err)
	}

	log.Println("\rList of accounts registered to current user: ")

	if c.FileData.Account.Name != "" {
		log.Printf("\r > %s - logged in        ", c.FileData.Account.Name)
	}
	for _, account := range accounts {
		log.Printf("\r > %s             ", account)
	}
	log.Println("\r                                      ")
	return nil
} 

//Returns the name of the currently logged in account
func commandGetCurrentAccount(c *config.Config, args []string) error {
	if len(args) != 0 {
		log.Println("\rCommand syntax issue - type help for more details")
		return nil 
	}

	if c.FileData.Account.Name == "" {
		log.Println("\rNo account currently logged in")
		return nil
	}

	acc := c.FileData.Account.Name

	log.Printf("\rAccount currently logged in: %s", acc)
	return nil
}

//Deletes a user's specified account
func commandDeleteAccount(c *config.Config, args []string) error {
	if len(args) != 1 {
		log.Println("\rCommand syntax issue - type help for more details")
		return nil 
	}

	getUrl := c.Client.BaseURL + "/api/users"

	username := c.FileData.User.Name
	accName := args[0]

	if c.FileData.Account.Name == "" {
		log.Println("\rNot currently logged into an account")
		return nil 
	}

	if accName != c.FileData.Account.Name {
		log.Println("\rAccount name given does not match currently logged in account")
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

		res.Body.Close()
		break
	}

	//If incorrect password maxtries was reached
	if attempts >= maxRetries {
		return fmt.Errorf("max password attempts reached")
	}

	scanner := bufio.NewScanner(os.Stdin)
	//Start another scanner loop for confirmation input
	for {
		fmt.Printf("Are you sure you want to delete account: %s? [y, n]\n", accName)
		fmt.Print(" > ")
		scanner.Scan()

		if scanner.Text() == "n" {
			fmt.Println("\rAccount deletion aborted")
			return nil
		} else if scanner.Text() != "y" {
			fmt.Println("\rPlease enter either [y] or [n]")
			continue
		} else {
			fmt.Printf("\rCommencing deletion of account: %s\n", accName)
			break
		}
	}

	deleteUrl := c.Client.BaseURL + "/api/users/" + c.FileData.User.ID.String() + "/accounts/" + c.FileData.Account.ID.String()

	log.Println("\rDeleting account's config files...")
	err = c.DeleteAccountDir()
	if err != nil {
		return err
	}
	log.Println("\rAccount's config files deleted successfully!")

	log.Println("\rDeleting account's database records...")
	response, err := c.MakeBasicRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return err 
	}
	defer response.Body.Close()

	if response.StatusCode > 400 {
		return fmt.Errorf("server returned error: %s", response.Status)
	}

	log.Println("\rAccount has been deleted successfully!")

	return nil
}