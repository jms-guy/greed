package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/config"
	"github.com/jms-guy/greed/models"
)

func commandCreateUser(c *config.Config, args []string) error {
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

	res, err := c.MakeBasicRequest("POST", registerURL, reqData)
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

	emailRes, err := c.MakeBasicRequest("POST", sendURL, emailData)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}

	if emailRes.StatusCode >= 400 {
		return fmt.Errorf("server returned error: %s", res.Status)
	}


	var code string

	fmt.Println(" < A verification email has been sent to the provided email address. > ")
	fmt.Println(" < Please enter the verification code provided within the email below. > ")
	for {
		fmt.Print(" > ")
		scanner.Scan()
		code = scanner.Text()

		if len(code) == 0 {
			fmt.Println(" < Please enter the verification code provided within the email below. > ")
			continue
		}
		break
	}
	verifyData := models.EmailVerificationWithCode{
		UserID: user.ID,
		Code: code,
	}

	verifyRes, err := c.MakeBasicRequest("POST", verifyURL, verifyData)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	
	if verifyRes.StatusCode >= 400 {
		return fmt.Errorf("server returned error: %s", verifyRes.Status)
	}

	
}