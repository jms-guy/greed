package cmd

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/cli/internal/utils"
	"github.com/jms-guy/greed/models"
)

// Helper for getting password while registering a new user
func registerPasswordHelper() (string, error) {
	var password string
	for {
		pw, err := auth.ReadPassword("Please enter a password > ")
		if err != nil {
			return "", fmt.Errorf("error getting password: %w", err)
		}

		if len(pw) < 8 {
			fmt.Println(" ")
			fmt.Println("Password must be greater than 8 characters")
			fmt.Println("")
			continue
		} else {
			for {
				confirmPw, err := auth.ReadPassword("Confirm password > ")
				if err != nil {
					return "", fmt.Errorf("error reading confirmed password: %w", err)
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

	return password, nil
}

// Helper for getting email while registering new user
func registerEmailHelper() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	var email string

	fmt.Println(" < Please enter a valid email address for your account. > ")
	fmt.Println(" < Email verification will be used for resetting your password, and recovering your account in case of a forgotten password. > ")
	fmt.Println(" < Type 'exit' to cancel > ")
	for {
		fmt.Print(" > ")
		scanner.Scan()
		email = scanner.Text()

		if len(email) == 0 {
			fmt.Println("Please enter an email address")
			continue
		}

		if email == "exit" {
			return "exit", nil
		}

		if auth.EmailValidation(email) {
			break
		}

		fmt.Println("Provided input is not a valid email address")
		continue
	}

	return email, nil
}

// Gets email code from user
func getEmailCodeHelper(app *CLIApp, userEmail, sendURL string, emailData models.EmailVerification) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	var code string

	fmt.Println(" < An email with a verification code has been sent to user's email. Please type in the code to continue > ")
	fmt.Println(" < Type 'exit' to cancel, or 'resend' to try again if no email has been received > ")
	for {
		fmt.Print(" > ")
		scanner.Scan()
		code = scanner.Text()

		if len(code) == 0 {
			fmt.Println(" < Please enter the verification code provided within the email below. > ")
			continue
		}

		if code == "exit" {
			return "", nil
		}

		if code == "resend" {
			emailRes, err := app.Config.MakeBasicRequest("POST", sendURL, "", emailData)
			if err != nil {
				return "", fmt.Errorf("error making http request: %w", err)
			}
			defer emailRes.Body.Close()

			err = checkResponseStatus(emailRes)
			if err != nil {
				return "", err
			}

			fmt.Printf(" < Another verification code has been sent to email: %s > \n", userEmail)
			continue
		}
		break
	}

	return code, nil
}

// Helper for performing email verification flow in registering a new user
func verifyEmailHelper(app *CLIApp, sendURL, verifyURL string, user models.User, emailData models.EmailVerification) (bool, error) {
	fmt.Println(" < It can take up to a couple of minutes for the email to be received > ")
	fmt.Println(" < Or you can type 'continue' to continue without a verified email address. > ")
	fmt.Println(" < Please note, if your address has not been verified, you will be unable to change your password, or recover your account in case the password has been forgotten. > ")

	code, err := getEmailCodeHelper(app, user.Email, sendURL, emailData)
	if err != nil {
		return false, err
	}
	if code == "" {
		return false, nil
	}

	if code == "continue" {
		params := database.CreateUserParams{
			ID:             user.ID.String(),
			Name:           user.Name,
			CreatedAt:      user.CreatedAt.Format("2006-01-02"),
			UpdatedAt:      user.UpdatedAt.Format("2006-01-02"),
			HashedPassword: user.HashedPassword,
			Email:          user.Email,
			IsVerified:     sql.NullBool{Bool: false, Valid: true},
		}

		_, err := app.Config.Db.CreateUser(context.Background(), params)
		if err != nil {
			return false, fmt.Errorf("error creating local record of user: %w", err)
		}

		fmt.Printf("User: %s has been successfully registered!\n", user.Name)
		fmt.Println("As a demo user, you have 10 total uses for commands (fetch, sync). The initial fetch will use 2, and each sync afterwards will also use 2.")
		return false, nil
	}

	verifyData := models.EmailVerificationWithCode{
		UserID: user.ID,
		Code:   code,
	}

	verifyRes, err := app.Config.MakeBasicRequest("POST", verifyURL, "", verifyData)
	if err != nil {
		return false, fmt.Errorf("error making http request: %w", err)
	}
	defer verifyRes.Body.Close()

	err = checkResponseStatus(verifyRes)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Helper for fetching user credentials on login
func userLoginHelper(app *CLIApp, username, loginURL string) (models.Credentials, error) {
	pw, err := auth.ReadPassword("Please enter your password > ")
	if err != nil {
		return models.Credentials{}, fmt.Errorf("error getting password: %w", err)
	}

	req := models.UserDetails{
		Name:     username,
		Password: pw,
	}

	res, err := app.Config.MakeBasicRequest("POST", loginURL, "", req)
	if err != nil {
		return models.Credentials{}, fmt.Errorf("error making request: %w", err)
	}
	defer res.Body.Close()

	err = checkResponseStatus(res)
	if err != nil {
		return models.Credentials{}, err
	}

	var login models.Credentials

	if err = json.NewDecoder(res.Body).Decode(&login); err != nil {
		return models.Credentials{}, fmt.Errorf("error decoding response data: %w", err)
	}

	return login, nil
}

// On user login, checks for existence of items, to know if this is a first login or not
func userCheckItemsHelper(app *CLIApp, login models.Credentials, itemsURL string) ([]models.ItemName, error) {
	var items []models.ItemName

	itemsRes, err := app.Config.MakeBasicRequest("GET", itemsURL, login.AccessToken, nil)
	if err != nil {
		return items, fmt.Errorf("error making request: %w", err)
	}
	defer itemsRes.Body.Close()

	err = checkResponseStatus(itemsRes)
	if err != nil {
		return items, err
	}

	if itemsRes.StatusCode == 200 {
		var itemsResp struct {
			Items []models.ItemName `json:"items"`
		}
		if err = json.NewDecoder(itemsRes.Body).Decode(&itemsResp); err != nil {
			return items, fmt.Errorf("error decoding response data: %w", err)
		}

		if len(itemsResp.Items) != 0 {
			items = append(items, itemsResp.Items...)

			fmt.Printf(" > Available items for user: %s\n", login.User.Name)
			fmt.Println(" ~~~~~")
			for _, i := range itemsResp.Items {
				fmt.Printf(" Institution: %s || Item Name: %s || ItemID: %s\n", i.InstitutionName, i.Nickname, i.ItemId)
			}
			fmt.Println("")

			err = auth.StoreTokens(login, app.Config.ConfigFP)
			if err != nil {
				return items, fmt.Errorf("error storing auth tokens: %w", err)
			}

			return items, nil
		}
	}

	return items, nil
}

// On user's first time loggin in, they will go through Plaid's Link flow, consisting of: asking server for
// Link token, opening browser link with token, getting a Public token from Plaid, and exchanging that public
// token with the server for a permanent access token
func userFirstTimePlaidLinkHelper(app *CLIApp, login models.Credentials, itemsURL string) (bool, error) {
	linkURL := app.Config.Client.BaseURL + "/plaid/get-link-token"
	redirectURL := app.Config.Client.BaseURL + "/link"
	accessURL := app.Config.Client.BaseURL + "/plaid/get-access-token"

	linkRes, err := app.Config.MakeBasicRequest("POST", linkURL, login.AccessToken, nil)
	if err != nil {
		return false, fmt.Errorf("error making request: %w", err)
	}
	defer linkRes.Body.Close()

	err = checkResponseStatus(linkRes)
	if err != nil {
		return false, err
	}

	var link models.LinkResponse
	if err = json.NewDecoder(linkRes.Body).Decode(&link); err != nil {
		return false, fmt.Errorf("error decoding response data: %w", err)
	}

	redirectURL = redirectURL + "?token=" + link.LinkToken
	err = utils.OpenLink(app.Config.OperatingSystem, redirectURL)
	if err != nil {
		return false, fmt.Errorf("error opening redirect link: %w", err)
	}

	fmt.Println("Waiting for Plaid callback...")
	token, err := auth.ListenForPlaidCallback()
	if err != nil {
		return false, fmt.Errorf("error getting public token: %w", err)
	}
	if token == "" {
		return false, fmt.Errorf("error getting public token")
	}

	name := promptForItemName()

	request := models.AccessTokenRequest{
		PublicToken: token,
		Nickname:    name,
	}

	// The following login.AccessTokens are app JWT's, not to be mistaken for Plaid's Access Tokens
	accessResponse, err := app.Config.MakeBasicRequest("POST", accessURL, login.AccessToken, request)
	if err != nil {
		return false, fmt.Errorf("error making request: %w", err)
	}
	defer accessResponse.Body.Close()

	err = checkResponseStatus(accessResponse)
	if err != nil {
		return false, err
	}

	itemsResp, err := app.Config.MakeBasicRequest("GET", itemsURL, login.AccessToken, nil)
	if err != nil {
		return false, fmt.Errorf("error making request: %w", err)
	}
	defer itemsResp.Body.Close()

	err = checkResponseStatus(itemsResp)
	if err != nil {
		return false, err
	}

	if itemsResp.StatusCode == 200 {
		var itemsResponse struct {
			Items []models.ItemName `json:"items"`
		}
		if err = json.NewDecoder(itemsResp.Body).Decode(&itemsResponse); err != nil {
			return false, fmt.Errorf("error decoding response data: %w", err)
		}

		if len(itemsResponse.Items) != 0 {
			fmt.Printf(" > Available items for user: %s\n", login.User.Name)
			fmt.Println(" ~~~~~")
			for _, i := range itemsResponse.Items {
				fmt.Printf(" Institution: %s || Item Name: %s || ItemID: %s\n", i.InstitutionName, i.Nickname, i.ItemId)
			}
			fmt.Println("")

			err = auth.StoreTokens(login, app.Config.ConfigFP)
			if err != nil {
				return false, fmt.Errorf("error storing auth tokens: %w", err)
			}

			return true, nil
		}
	}

	return true, nil
}

// Simple bufio scanner for getting an item nickname
func promptForItemName() string {
	scanner := bufio.NewScanner(os.Stdin)
	var name string

Loop:
	for {
		fmt.Println("Please enter a name for this item: ")
		fmt.Print(" > ")
		scanner.Scan()

		name = scanner.Text()

		if name == "" {
			fmt.Println("No input found")
			continue
		}

		fmt.Printf("Set item name to %s? (y/n)\n", name)
		for {
			scanner.Scan()
			confirm := scanner.Text()

			if confirm == "y" {
				break Loop
			} else if confirm == "n" {
				break
			} else {
				continue
			}
		}
	}

	return name
}

// Fetches webhook records from server, and searches for actions that user must take to resolve them
func checkForWebhookRecords(app *CLIApp, webhookURL string, items []models.ItemName) error {
	res, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", webhookURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	err = checkResponseStatus(res)
	if err != nil {
		return err
	}

	var webhookRecords []models.WebhookRecord
	if err = json.NewDecoder(res.Body).Decode(&webhookRecords); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	if len(webhookRecords) == 0 {
		return nil
	}

	itemsMap := make(map[string]string)
	for _, item := range items {
		itemsMap[item.ItemId] = item.Nickname
	}

	loginRequired := false
	syncRequired := false

	for _, record := range webhookRecords {
		_, found := itemsMap[record.ItemID]
		if !found {
			continue
		}

		switch record.WebhookCode {
		case "ITEM_LOGIN_REQUIRED", "ITEM_ERROR", "ITEM_BAD_STATE", "NEW_ACCOUNTS_AVAILABLE", "PENDING_DISCONNECT":
			loginRequired = true
		case "TRANSACTIONS_UPDATES_AVAILABLE", "DEFAULT_UPDATE", "TRANSACTIONS_REMOVED":
			if !loginRequired {
				syncRequired = true
			}
		default:
		}
	}

	if loginRequired {
		fmt.Println("One or more of your bank connections require re-authentication. Please use the 'update' command.")
	} else if syncRequired {
		fmt.Println("New data is available for one or more accounts. Please use the 'sync <item-name>' command.")
	}

	return nil
}
