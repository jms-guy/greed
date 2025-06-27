package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	//"github.com/jms-guy/greed/cli/internal/utils"
	"github.com/google/uuid"
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
	defer emailRes.Body.Close()

	if emailRes.StatusCode >= 400 {
		return fmt.Errorf("server returned error: %s", emailRes.Status)
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
	defer verifyRes.Body.Close()
	
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
//Currently using sandbox flow
func commandUserLogin(c *config.Config, args []string) error {
	if len(args) != 1 {
		fmt.Printf("Incorrect number of arguments given - type --help for more details")
		return nil 
	}

	username := args[0]

	loginURL := c.Client.BaseURL + "/api/auth/login"
	itemsURL := c.Client.BaseURL + "/api/items"
	linkURL := c.Client.BaseURL + "/plaid/get-link-token"
	//redirectURL := c.Client.BaseURL + "/link"
	sandboxURL := c.Client.BaseURL + "/admin/sandbox"

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
		fmt.Printf("Bad request - %s\n", res.Status)
		return nil
	}

	var login models.Credentials

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

	/*
	redirectURL = redirectURL + "?token=" + link.LinkToken
	err = utils.OpenLink(c.OperatingSystem, redirectURL)
	if err != nil {
		return fmt.Errorf("error opening redirect link: %w", err)
	}
	*/

	tokenRes, err := c.MakeBasicRequest("POST", sandboxURL, login.AccessToken, nil)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer tokenRes.Body.Close()

	if tokenRes.StatusCode > 500 {
		fmt.Println("Server error")
		return nil 
	}

	if tokenRes.StatusCode >= 400 {
		fmt.Printf("Request failed with code %d\n", tokenRes.StatusCode)
		body, _ := io.ReadAll(tokenRes.Body)
		fmt.Println("Response body:", string(body))
		return nil
	}
	fmt.Println("Item record created!")

	itemsResp, err := c.MakeBasicRequest("GET", itemsURL, login.AccessToken, nil)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer itemsResp.Body.Close()

	if itemsResp.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil 
	}

	if itemsResp.StatusCode == 200 {
		var itemsResponse struct {
			Items []models.ItemName `json:"items"`
		}
		if err = json.NewDecoder(itemsResp.Body).Decode(&itemsResponse); err != nil {
			return fmt.Errorf("error decoding response data: %w", err)
		}
		
		if len(itemsResponse.Items) != 0 {
			fmt.Printf(" > Available items for user: %s\n", username)
			fmt.Println(" ~~~~~")
			for _, i := range itemsResponse.Items {
				fmt.Printf(" Institution: %s || Item Name: %s || ItemID: %s\n", i.InstitutionName, i.Nickname, i.ItemId)
			}
			return nil
		}
	
	}

	return nil
}

//Logs a user out, by deleting their local credentials file, and expiring their session delegation server side
func commandUserLogout(c *config.Config, args []string) error {
	if len(args) > 0 {
		fmt.Println("Incorrect number of arguments given - type --help for more details")
		return nil 
	}

	logoutURL := c.Client.BaseURL + "/api/auth/logout"

	creds, err := auth.GetCreds(c.ConfigFP)
	if err != nil {
		fmt.Printf("File error: %s\n", err)
		return nil
	}

	request := models.RefreshRequest{
		RefreshToken: creds.RefreshToken,
	}

	resp, err := c.MakeBasicRequest("POST", logoutURL, "", request)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil 
	}

	if resp.StatusCode >= 400 {
		fmt.Printf("Bad request - %s\n", resp.Status)
		return nil 
	}

	err = auth.RemoveCreds(c.ConfigFP)
	if err != nil {
		fmt.Printf("Error logging out - %s\n", err)
		return nil 
	}

	fmt.Println("Logged out successfully!")
	return nil
}

//Delete a user's records locally, and server side
func commandDeleteUser(c *config.Config, args []string) error {
	if len(args) != 1 {
		fmt.Println("Incorrect number of arguments given - type --help for more details")
		return nil
	}

	username := args[0]

	deleteURL := c.Client.BaseURL + "/api/users/me"

	pw, err := auth.ReadPassword("Please enter your password > ")
	if err != nil {
		return fmt.Errorf("error getting password: %w", err)
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println(" < Upon completion, all records, both local and server side, will be permanently deleted. > ")
	fmt.Printf(" < Are you sure you want to delete user - %s? (y/n) > \n", username)
	for {
		fmt.Print(" > ")
		scanner.Scan()
		if scanner.Text() == "n" {
			fmt.Println("User deletion aborted.")
			return nil 
		} else if scanner.Text() == "y" {
			break
		} else {
			fmt.Println(" < Please enter either 'y' or 'n' > ")
		}
	}

	user, err := c.Db.GetUser(context.Background(), username)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No user found on machine with that name")
			return nil 
		}
		return fmt.Errorf("error getting user from local database: %w", err)
	}

	err = auth.ValidatePasswordHash(user.HashedPassword, pw)
	if err != nil {
		fmt.Println("Incorrect password entered")
		return nil
	}

	resp, err := DoWithAutoRefresh(c, func(token string) (*http.Response, error) {
		return c.MakeBasicRequest("DELETE", deleteURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil 
	}

	if resp.StatusCode >= 400 {
		fmt.Printf("Bad request - %s\n", resp.Status)
		return nil 
	}

	err = c.Db.DeleteUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("error deleting local user record: %w", err)
	}

	err = auth.RemoveCreds(c.ConfigFP)
	if err != nil {
		fmt.Printf("Error removing credentials - %s\n", err)
		return nil 
	}

	fmt.Printf("All records for %s have been successfully deleted\n", username)
	return nil
}

//Verifies a user's email address
func commandVerifyEmail(c *config.Config, args []string) error {
	if len(args) != 0 {
		fmt.Println("Incorrect number of arguments - type --help for more details")
		return nil 
	}

	sendURL := c.Client.BaseURL + "/api/auth/email/send"
	verifyURL := c.Client.BaseURL + "/api/auth/email/verify"

	creds, err := auth.GetCreds(c.ConfigFP)
	if err != nil {
		fmt.Printf("Error getting credentials - %s\n", err)
		return nil
	}

	user, err := c.Db.GetUser(context.Background(), creds.User.Name)
	if err != nil {
		return fmt.Errorf("error getting local user record: %w", err)
	}

	if user.IsVerified.Bool {
		fmt.Println("User already has a verified email address")
		return nil
	}

	sendReq := models.EmailVerification{
		UserID: creds.User.ID,
		Email: user.Email,
	}

	sendResp, err := c.MakeBasicRequest("POST", sendURL, "", sendReq)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer sendResp.Body.Close()

	if sendResp.StatusCode >= 400 {
		return fmt.Errorf("server returned error: %s", sendResp.Status)
	}

	scanner := bufio.NewScanner(os.Stdin)
	var code string

	fmt.Println(" < A verification email has been sent to the provided email address. > ")
	fmt.Println(" < Please enter the verification code provided within the email below. > ")
	fmt.Println(" < It can take up to a couple of minutes for the email to be received. If no email has been received, you can type 'resend' to try again. > ")
	for {
		fmt.Print(" > ")
		scanner.Scan()
		code = scanner.Text()

		if len(code) == 0 {
			fmt.Println(" < Please enter the verification code provided within the email below. > ")
			continue
		}

		if code == "resend" {
			emailRes, err := c.MakeBasicRequest("POST", sendURL, "", sendReq)
			if err != nil {
				return fmt.Errorf("error making http request: %w", err)
			}

			if emailRes.StatusCode >= 400 {
				return fmt.Errorf("server returned error: %s", emailRes.Status)
			}

			fmt.Printf(" < Another verification code has been sent to email: %s > \n", user.Email)
			continue
		}
		break
	}
	verifyData := models.EmailVerificationWithCode{
		UserID: creds.User.ID,
		Code: code,
	}

	verifyRes, err := c.MakeBasicRequest("POST", verifyURL, "", verifyData)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer verifyRes.Body.Close()
	
	if verifyRes.StatusCode >= 400 {
		return fmt.Errorf("server returned error: %s", verifyRes.Status)
	}

	verifyParams := database.VerifyEmailParams{
		IsVerified: sql.NullBool{Bool: true, Valid: true},
		UpdatedAt: time.Now().Format("2006-01-02"),
		Name: user.Name,
	}

	err = c.Db.VerifyEmail(context.Background(), verifyParams)
	if err != nil {
		return fmt.Errorf("error verifying email in user record: %w", err)
	}

	fmt.Printf("User '%s' email '%s' verified successfully\n", user.Name, user.Email)

	return nil
}

//Lists a user's items
func commandUserItems(c *config.Config, args []string) error {
	if len(args) != 0 {
		fmt.Println("Incorrect number of arguments - type --help for more details")
		return nil 
	}

	itemsURL := c.Client.BaseURL + "/api/items"

	creds, err := auth.GetCreds(c.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	resp, err := DoWithAutoRefresh(c, func(token string) (*http.Response, error) {
		return c.MakeBasicRequest("GET", itemsURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil 
	}

	if resp.StatusCode == 200 {
		var itemsResponse struct {
			Items []models.ItemName `json:"items"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&itemsResponse); err != nil {
			return fmt.Errorf("error decoding response data: %w", err)
		}
		
		if len(itemsResponse.Items) != 0 {
			fmt.Printf(" > Available items for user: %s\n", creds.User.Name)
			fmt.Println(" ~~~~~")
			for _, i := range itemsResponse.Items {
				fmt.Printf(" Institution: %s || Item Name: %s || ItemID: %s\n", i.InstitutionName, i.Nickname, i.ItemId)
			}
			return nil
		}
	}
	return nil
}

//Updates a user's password
func commandUpdatePassword(c *config.Config, args []string) error {
	if len(args) != 0 {
		fmt.Println("Incorrect number of arguments - type --help for more details")
		return nil 
	}

	sendURL := c.Client.BaseURL + "/api/auth/email/send"
	updateURL := c.Client.BaseURL + "/api/users/update-password"

	creds, err := auth.GetCreds(c.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	user, err := c.Db.GetUser(context.Background(), creds.User.Name)
	if err != nil {
		return fmt.Errorf("error getting local user record: %w", err)
	}

	pw, err := auth.ReadPassword("Please enter your current password > ")
	if err != nil {
		return fmt.Errorf("error getting password: %w", err)
	}

	err = auth.ValidatePasswordHash(user.HashedPassword, pw)
	if err != nil {
		fmt.Println("Bad password input")
		return nil
	}

	var password string

	for {
		pw, err := auth.ReadPassword("Please enter a new password for your user > ")
		if err != nil {
			return fmt.Errorf("error getting password: %w", err)
		}

		if len(pw) < 8 {
			fmt.Println("Password must be greater than 8 characters")
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

	request := models.EmailVerification{
		UserID: creds.User.ID,
		Email: creds.User.Email,
	}

	emailResp, err := c.MakeBasicRequest("POST", sendURL, "", request)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer emailResp.Body.Close()

	if emailResp.StatusCode >= 500 {
		fmt.Println("Server error")
	}

	scanner := bufio.NewScanner(os.Stdin)
	var code string

	fmt.Println(" < An email with a verification code has been sent to user's email. Please type in the code to continue > ")
		for {
		fmt.Print(" > ")
		scanner.Scan()
		code = scanner.Text()

		if len(code) == 0 {
			fmt.Println(" < Please enter the verification code provided within the email below. > ")
			continue
		}

		if code == "resend" {
			emailRes, err := c.MakeBasicRequest("POST", sendURL, "", request)
			if err != nil {
				return fmt.Errorf("error making http request: %w", err)
			}

			if emailRes.StatusCode >= 400 {
				return fmt.Errorf("server returned error: %s", emailRes.Status)
			}

			fmt.Printf(" < Another verification code has been sent to email: %s > \n", creds.User.Email)
			continue
		}
		break
	}

	updateReq := models.UpdatePassword{
		NewPassword: password,
		Code: code,
	}

	resp, err := DoWithAutoRefresh(c, func(token string) (*http.Response, error) {
		return c.MakeBasicRequest("PUT", updateURL, token, updateReq)
	})
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil 
	}

	var updated models.UpdatedPassword

	if err = json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	params := database.UpdatePasswordParams{
		HashedPassword: updated.HashPassword,
		UpdatedAt: time.Now().Format("2006-01-02"),
		Name: creds.User.Name,
	}

	err = c.Db.UpdatePassword(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error updating local user record: %w", err)
	}

	fmt.Println("Password updated successfully")
	return nil
}

//Resets a user's forgotten password, allowing for account recovery
func commandResetPassword(c *config.Config, args []string) error {
	if len(args) != 1 {
		fmt.Println("Incorrect number of arguments - type --help for more details")
		return nil 
	}

	email := args[0]

	sendURL := c.Client.BaseURL + "/api/auth/email/send"
	resetURL := c.Client.BaseURL + "/api/auth/reset-password"

	user, err := c.Db.GetUserByEmail(context.Background(), email)
	if err != nil {
		return fmt.Errorf("error getting user record: %w", err)
	}

	var password string

	for {
		pw, err := auth.ReadPassword("Please enter a new password for your user > ")
		if err != nil {
			return fmt.Errorf("error getting password: %w", err)
		}

		if len(pw) < 8 {
			fmt.Println("Password must be greater than 8 characters")
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

	request := models.EmailVerification{
		UserID: uuid.MustParse(user.ID),
		Email: user.Email,
	}

	emailResp, err := c.MakeBasicRequest("POST", sendURL, "", request)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer emailResp.Body.Close()

	if emailResp.StatusCode >= 500 {
		fmt.Println("Server error")
	}

	scanner := bufio.NewScanner(os.Stdin)
	var code string

	fmt.Println(" < An email with a verification code has been sent to user's email. Please type in the code to continue > ")
		for {
		fmt.Print(" > ")
		scanner.Scan()
		code = scanner.Text()

		if len(code) == 0 {
			fmt.Println(" < Please enter the verification code provided within the email below. > ")
			continue
		}

		if code == "resend" {
			emailRes, err := c.MakeBasicRequest("POST", sendURL, "", request)
			if err != nil {
				return fmt.Errorf("error making http request: %w", err)
			}

			if emailRes.StatusCode >= 400 {
				return fmt.Errorf("server returned error: %s", emailRes.Status)
			}

			fmt.Printf(" < Another verification code has been sent to email: %s > \n",user.Email)
			continue
		}
		break
	}

	resetReq := models.ResetPassword{
		Email: user.Email,
		Code: code,
		NewPassword: password,
	}

	resetResp, err := c.MakeBasicRequest("POST", resetURL, "", resetReq)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer resetResp.Body.Close()

	if resetResp.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil 
	}

	var updated models.UpdatedPassword

	if err = json.NewDecoder(resetResp.Body).Decode(&updated); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	params := database.UpdatePasswordParams{
		HashedPassword: updated.HashPassword,
		UpdatedAt: time.Now().Format("2006-01-02"),
		Name: user.Name,
	}

	err = c.Db.UpdatePassword(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error updating local user record: %w", err)
	}

	fmt.Println("Password successfully reset")
	return nil
}