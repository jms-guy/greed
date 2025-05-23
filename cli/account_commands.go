package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

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

	for _, account := range c.FileData.Accounts {
		if account.Name == accountName {
			log.Printf("\rUser already has an account named %s\n", accountName)
			return nil
		}
	}

	userId := c.FileData.User.ID

	//Set api endpoint url
	url := c.Client.BaseURL + "/api/users/" + userId.String() + "/accounts"

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
		InputType: inputType,
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

	c.FileData.Accounts = append(c.FileData.Accounts, account)

	//Save updated FileData to the user config file in the currentsession directory
	err = c.SaveFileData()
	if err != nil {
		return fmt.Errorf("error saving file data: %w", err)
	}

	log.Println("\rAccount created successfully!")

	return nil
}