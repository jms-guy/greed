package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jms-guy/greed/cli/internal/utils"
	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/config"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/models"
)

//Function registers a new user with the server
func commandRegisterUser(c *config.Config, args []string) error {
	if len(args) != 1 {
		fmt.Printf("Incorrect number of arguments given - type --help for more details")
		return nil 
	}

	username := args[0]

	registerURL := c.Client.BaseURL + "/api/auth/register"
	sendURL := c.Client.BaseURL + "/api/auth/email/send"
	verifyURL := c.Client.BaseURL + "/api/auth/email/verify"

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

	scanner := bufio.NewScanner(os.Stdin)
	var email string

	fmt.Println(" < Please enter a valid email address for your account. > ")
	fmt.Println(" < Email verification will be used for resetting your password, and recovering your account in case of a forgotten password. > ")
	for {
		fmt.Print(" > ")
		scanner.Scan()
		email = scanner.Text()

		if len(email) == 0 {
			fmt.Println("Please enter an email address")
			continue
		}
		
		if auth.EmailValidation(email) {
			break
		}

		fmt.Println("Provided input is not a valid email address")
		continue
	}

	reqData := models.UserDetails{
		Name: username,
		Password: password,
		Email: email,
	}

	res, err := c.MakeBasicRequest("POST", registerURL, "", reqData)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("server returned error: %s", res.Status)
	}

	var user models.User

	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	res.Body.Close()

	emailData := models.EmailVerification{
		UserID: user.ID,
		Email: user.Email,
	}

	emailRes, err := c.MakeBasicRequest("POST", sendURL, "", emailData)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}

	if emailRes.StatusCode >= 400 {
		return fmt.Errorf("server returned error: %s", res.Status)
	}


	var code string

	fmt.Println(" < A verification email has been sent to the provided email address. > ")
	fmt.Println(" < Please enter the verification code provided within the email below. > ")
	fmt.Println(" < It can take up to a couple of minutes for the email to be received. If no email has been received, you can type 'resend' to try again. > ")
	fmt.Println(" < Or you can type 'continue' to continue without a verified email address. > ")
	fmt.Println(" < Please note, if your address has not been verified, you will be unable to change your password, or recover your account in case the password has been forgotten. > ")
	for {
		fmt.Print(" > ")
		scanner.Scan()
		code = scanner.Text()

		if len(code) == 0 {
			fmt.Println(" < Please enter the verification code provided within the email below. > ")
			continue
		}

		if code == "resend" {
			emailRes, err := c.MakeBasicRequest("POST", sendURL, "", emailData)
			if err != nil {
				return fmt.Errorf("error making http request: %w", err)
			}

			if emailRes.StatusCode >= 400 {
				return fmt.Errorf("server returned error: %s", res.Status)
			}

			fmt.Printf(" < Another verification code has been sent to email: %s > \n", email)
			continue
		}

		if code == "continue" {
			params := database.CreateUserParams{
				ID: user.ID.String(),
				Name: username,
				CreatedAt: user.CreatedAt.Format("2006-01-02"),
				UpdatedAt: user.UpdatedAt.Format("2006-01-02"),
				HashedPassword: user.HashedPassword,
				Email: user.Email,
				IsVerified: sql.NullBool{Bool: false, Valid: true},
			}
		
			_, err = c.Db.CreateUser(context.Background(), params)
			if err != nil {
				return fmt.Errorf("error creating local record of user: %w", err)	
			}
		
			fmt.Printf("User: %s has been successfully registered!\n", username)
			return nil
		}
		break
	}
	verifyData := models.EmailVerificationWithCode{
		UserID: user.ID,
		Code: code,
	}

	verifyRes, err := c.MakeBasicRequest("POST", verifyURL, "", verifyData)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	
	if verifyRes.StatusCode >= 400 {
		return fmt.Errorf("server returned error: %s", verifyRes.Status)
	}

	params := database.CreateUserParams{
		ID: user.ID.String(),
		Name: username,
		CreatedAt: user.CreatedAt.Format("2006-01-02"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02"),
		HashedPassword: user.HashedPassword,
		Email: user.Email,
		IsVerified: sql.NullBool{Bool: true, Valid: true},
	}

	_, err = c.Db.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error creating local record of user: %w", err)	
	}

	fmt.Printf("User: %s has been successfully registered!\n", username)
	return nil
}

//Function creates a login session for user, getting auth tokens
func commandUserLogin(c *config.Config, args []string) error {
	if len(args) != 1 {
		fmt.Printf("Incorrect number of arguments given - type --help for more details")
		return nil 
	}

	username := args[0]

	loginURL := c.Client.BaseURL + "/api/auth/login"
	itemsURL := c.Client.BaseURL + "/api/items"
	linkURL := c.Client.BaseURL + "/plaid/get-link-token"
	redirectURL := c.Client.BaseURL + "/link"

	pw, err := auth.ReadPassword("Please enter your password > ")
	if err != nil {
		return fmt.Errorf("error getting password: %w", err)
	}

	req := models.UserDetails{
		Name: username,
		Password: pw,
	}

	res, err := c.MakeBasicRequest("POST", loginURL, "", req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil
	}
	if res.StatusCode >= 400 {
		fmt.Printf("Bad request - %s", res.Status)
		return nil
	}

	var login models.LoginResponse

	if err = json.NewDecoder(res.Body).Decode(&login); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	err = auth.StoreTokens(login, c.ConfigFP)
	if err != nil {
		return fmt.Errorf("error storing auth tokens: %w", err)
	}

	itemsRes, err := c.MakeBasicRequest("GET", itemsURL, login.AccessToken, nil)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer itemsRes.Body.Close()

	if itemsRes.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil 
	}

	if itemsRes.StatusCode == 200 {
		var itemsResp struct {
			Items []models.ItemName `json:"items"`
		}
		if err = json.NewDecoder(itemsRes.Body).Decode(&itemsResp); err != nil {
			return fmt.Errorf("error decoding response data: %w", err)
		}
		
		if len(itemsResp.Items) != 0 {
			fmt.Printf(" > Available items for user: %s\n", username)
			fmt.Println(" ~~~~~")
			for _, i := range itemsResp.Items {
				fmt.Printf(" Institution: %s || Item Name: %s || ItemID: %s\n", i.InstitutionName, i.Nickname, i.ItemId)
			}
			return nil
		}
	
	}
	
	linkRes, err := c.MakeBasicRequest("POST", linkURL, login.AccessToken, nil)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer linkRes.Body.Close()

	if linkRes.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil 
	}

	if linkRes.StatusCode >= 400 {
		fmt.Printf("Bad request - %s", linkRes.Status)
		return nil
	}
	
	var link models.LinkResponse
	if err = json.NewDecoder(linkRes.Body).Decode(&link); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	err = utils.OpenLink(c.OperatingSystem, redirectURL)
	if err != nil {
		return fmt.Errorf("error opening redirect link: %w", err)
	}

	return nil
}

func commandUserLogout(c *config.Config, args []string) error {
	return nil 
}

func commandDeleteUser(c *config.Config, args []string) error {
	return nil
}