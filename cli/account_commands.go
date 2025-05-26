package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"github.com/jms-guy/greed/cli/internal/config"
	"github.com/jms-guy/greed/models"
)

//Function will create a new account for user
func commandCreateAccount(c *config.Config, args []string) error {
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
		*/

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
	}

	//Create account request
	reqParams := models.CreateAccount{
		Name: accountName,
		Currency: currencyType,
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

func commandAccountLogin(c *config.Config, args []string) error {
	if len(args) != 1 {
		log.Println("\rCommand syntax issue - type help for more details")
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