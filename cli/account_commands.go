package main

import (
	"bufio"
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
			log.Printf("User already has an account named %s\n", accountName)
		}
	}

	userId := c.FileData.User.ID

	//Set api endpoint url
	url := c.Client.BaseURL + "/api/users/" + userId.String() + "/accounts"

	log.Println("\rCreating account:               ")
	log.Println("\r                             ")
	log.Println("\rWill this account be synced to a financial institution, or will the data be manually input? ")

	//Start bufio scanner for additional account fields(InputType and Currency)
	inputType := ""
	currencyType := ""

	scanner := bufio.NewScanner(os.Stdin)
	for {
		log.Printf("\r[auto]/[manual] >      ")
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

	log.Println("\rPlease set the base currency for this account. May be changed at later.")
	log.Println("\rIf no currency enteredm will be defaulted to 'CAD'. For a list of supported currencies, type")
	log.Println("\r [currencies]                        ")

	for {
		log.Printf("\r Currency >      ")
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


}