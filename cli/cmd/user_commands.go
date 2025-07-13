package cmd

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/cli/internal/utils"
	"github.com/jms-guy/greed/models"
)

//Function creates a new user record on server, and in loca database
func (app *CLIApp) commandRegisterUser(args []string) error {

	username := args[0]

	registerURL := app.Config.Client.BaseURL + "/api/auth/register"
	sendURL := app.Config.Client.BaseURL + "/api/auth/email/send"
	verifyURL := app.Config.Client.BaseURL + "/api/auth/email/verify"

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
			return nil
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

	res, err := app.Config.MakeBasicRequest("POST", registerURL, "", reqData)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	checkResponseStatus(res)

	var user models.User

	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	emailData := models.EmailVerification{
		UserID: user.ID,
		Email: user.Email,
	}

	emailRes, err := app.Config.MakeBasicRequest("POST", sendURL, "", emailData)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer emailRes.Body.Close()

	checkResponseStatus(emailRes)

	var code string

	fmt.Println(" < A verification email has been sent to the provided email address. > ")
	fmt.Println(" < Please enter the verification code provided within the email below. > ")
	fmt.Println(" < It can take up to a couple of minutes for the email to be received. If no email has been received, you can type 'resend' to try again. > ")
	fmt.Println(" < Or you can type 'continue' to continue without a verified email address. > ")
	fmt.Println(" < Please note, if your address has not been verified, you will be unable to change your password, or recover your account in case the password has been forgotten. > ")
	fmt.Println(" < Type 'exit' to cancel > ")
	for {
		fmt.Print(" > ")
		scanner.Scan()
		code = scanner.Text()

		if len(code) == 0 {
			fmt.Println(" < Please enter the verification code provided within the email below. > ")
			continue
		}

		if code == "exit" {
			return nil
		}

		if code == "resend" {
			emailRes, err := app.Config.MakeBasicRequest("POST", sendURL, "", emailData)
			if err != nil {
				return fmt.Errorf("error making http request: %w", err)
			}
			defer emailRes.Body.Close()

			checkResponseStatus(emailRes)

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
		
			_, err = app.Config.Db.CreateUser(context.Background(), params)
			if err != nil {
				return fmt.Errorf("error creating local record of user: %w", err)	
			}
		
			fmt.Printf("User: %s has been successfully registered!\n", username)
			fmt.Println("As a demo user, you have 10 total uses for commands (fetch, sync). The intial fetch will use 2, and each sync afterwards will also use 2.")
			return nil
		}
		break
	}
	verifyData := models.EmailVerificationWithCode{
		UserID: user.ID,
		Code: code,
	}

	verifyRes, err := app.Config.MakeBasicRequest("POST", verifyURL, "", verifyData)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer verifyRes.Body.Close()
	
	checkResponseStatus(verifyRes)

	params := database.CreateUserParams{
		ID: user.ID.String(),
		Name: username,
		CreatedAt: user.CreatedAt.Format("2006-01-02"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02"),
		HashedPassword: user.HashedPassword,
		Email: user.Email,
		IsVerified: sql.NullBool{Bool: true, Valid: true},
	}

	_, err = app.Config.Db.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error creating local record of user: %w", err)	
	}

	fmt.Printf("User: %s has been successfully registered!\n", username)
	fmt.Println("As a demo user, you have 10 total uses for commands (fetch, sync). The intial fetch will use 2, and each sync afterwards will also use 2.")
	
	return nil
}

//Function creates a login session for user, getting auth tokens
func (app *CLIApp) commandUserLogin(args []string) error {

	username := args[0]

	loginURL := app.Config.Client.BaseURL + "/api/auth/login"
	itemsURL := app.Config.Client.BaseURL + "/api/items"
	linkURL := app.Config.Client.BaseURL + "/plaid/get-link-token"
	redirectURL := app.Config.Client.BaseURL + "/link"
	accessURL := app.Config.Client.BaseURL + "/plaid/get-access-token"

	pw, err := auth.ReadPassword("Please enter your password > ")
	if err != nil {
		return fmt.Errorf("error getting password: %w", err)
	}

	req := models.UserDetails{
		Name: username,
		Password: pw,
	}

	res, err := app.Config.MakeBasicRequest("POST", loginURL, "", req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer res.Body.Close()

	checkResponseStatus(res)

	var login models.Credentials

	if err = json.NewDecoder(res.Body).Decode(&login); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	itemsRes, err := app.Config.MakeBasicRequest("GET", itemsURL, login.AccessToken, nil)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer itemsRes.Body.Close()

	checkResponseStatus(itemsRes)

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

			err = auth.StoreTokens(login, app.Config.ConfigFP)
			if err != nil {
				return fmt.Errorf("error storing auth tokens: %w", err)
			}
			
			return nil
		}
	}
	
	linkRes, err := app.Config.MakeBasicRequest("POST", linkURL, login.AccessToken, nil)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer linkRes.Body.Close()

	checkResponseStatus(linkRes)
	
	var link models.LinkResponse
	if err = json.NewDecoder(linkRes.Body).Decode(&link); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	redirectURL = redirectURL + "?token=" + link.LinkToken
	err = utils.OpenLink(app.Config.OperatingSystem, redirectURL)
	if err != nil {
		return fmt.Errorf("error opening redirect link: %w", err)
	}

	fmt.Println("Waiting for Plaid callback...")
	token, err := auth.ListenForPlaidCallback()
	if err != nil {
		return fmt.Errorf("error getting public token: %w", err)
	}
	if token == "" {
		return fmt.Errorf("error getting public token")
	}

	name := promptForItemName()

	request := models.AccessTokenRequest{
		PublicToken: token,
		Nickname: name,
	}
	accessResponse, err := app.Config.MakeBasicRequest("POST", accessURL, login.AccessToken, request)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer accessResponse.Body.Close()

	checkResponseStatus(accessResponse)

	itemsResp, err := app.Config.MakeBasicRequest("GET", itemsURL, login.AccessToken, nil)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer itemsResp.Body.Close()

	checkResponseStatus(itemsResp)

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
			
			err = auth.StoreTokens(login, app.Config.ConfigFP)
			if err != nil {
				return fmt.Errorf("error storing auth tokens: %w", err)
			}

			return nil
		}
	
	}

	
	err = auth.StoreTokens(login, app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error storing auth tokens: %w", err)
	}

	return nil
}

//Logs a user out, by deleting their local credentials file, and expiring their session delegation server side
func (app *CLIApp) commandUserLogout() error {

	logoutURL := app.Config.Client.BaseURL + "/api/auth/logout"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		fmt.Printf("File error: %s\n", err)
		return nil
	}

	request := models.RefreshRequest{
		RefreshToken: creds.RefreshToken,
	}

	resp, err := app.Config.MakeBasicRequest("POST", logoutURL, "", request)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	checkResponseStatus(resp)

	err = auth.RemoveCreds(app.Config.ConfigFP)
	if err != nil {
		fmt.Printf("Error logging out - %s\n", err)
		return nil 
	}

	fmt.Println("Logged out successfully!")
	return nil
}

//Delete a user's records locally, and server side
func (app *CLIApp) commandDeleteUser(args []string) error {

	username := args[0]

	deleteURL := app.Config.Client.BaseURL + "/api/users/me"

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

	user, err := app.Config.Db.GetUser(context.Background(), username)
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

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("DELETE", deleteURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	checkResponseStatus(resp)

	err = app.Config.Db.DeleteUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("error deleting local user record: %w", err)
	}

	err = auth.RemoveCreds(app.Config.ConfigFP)
	if err != nil {
		fmt.Printf("Error removing credentials - %s\n", err)
		return nil 
	}

	fmt.Printf("All records for %s have been successfully deleted\n", username)
	return nil
}

//Verifies a user's email address
func (app *CLIApp) commandVerifyEmail() error {

	sendURL := app.Config.Client.BaseURL + "/api/auth/email/send"
	verifyURL := app.Config.Client.BaseURL + "/api/auth/email/verify"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		fmt.Printf("Error getting credentials - %s\n", err)
		return nil
	}

	user, err := app.Config.Db.GetUser(context.Background(), creds.User.Name)
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

	sendResp, err := app.Config.MakeBasicRequest("POST", sendURL, "", sendReq)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer sendResp.Body.Close()

	checkResponseStatus(sendResp)

	scanner := bufio.NewScanner(os.Stdin)
	var code string

	fmt.Println(" < A verification email has been sent to the provided email address. > ")
	fmt.Println(" < Please enter the verification code provided within the email below. > ")
	fmt.Println(" < It can take up to a couple of minutes for the email to be received. If no email has been received, you can type 'resend' to try again. > ")
	fmt.Println(" < Type 'exit' to cancel > ")
	for {
		fmt.Print(" > ")
		scanner.Scan()
		code = scanner.Text()

		if len(code) == 0 {
			fmt.Println(" < Please enter the verification code provided within the email below. > ")
			continue
		}

		if code == "exit" {
			return nil
		}

		if code == "resend" {
			emailRes, err := app.Config.MakeBasicRequest("POST", sendURL, "", sendReq)
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

	verifyRes, err := app.Config.MakeBasicRequest("POST", verifyURL, "", verifyData)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer verifyRes.Body.Close()
	
	checkResponseStatus(verifyRes)

	verifyParams := database.VerifyEmailParams{
		IsVerified: sql.NullBool{Bool: true, Valid: true},
		UpdatedAt: time.Now().Format("2006-01-02"),
		Name: user.Name,
	}

	err = app.Config.Db.VerifyEmail(context.Background(), verifyParams)
	if err != nil {
		return fmt.Errorf("error verifying email in user record: %w", err)
	}

	fmt.Printf("User '%s' email '%s' verified successfully\n", user.Name, user.Email)

	return nil
}

//Lists all items attached to a specific user
func (app *CLIApp) commandUserItems() error {

	itemsURL := app.Config.Client.BaseURL + "/api/items"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", itemsURL, token, nil)
	})
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	checkResponseStatus(resp)
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

//Updates a user's password in record. Requires verified email address to send code to
func (app *CLIApp) commandUpdatePassword() error {

	sendURL := app.Config.Client.BaseURL + "/api/auth/email/send"
	updateURL := app.Config.Client.BaseURL + "/api/users/update-password"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	user, err := app.Config.Db.GetUser(context.Background(), creds.User.Name)
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

	emailResp, err := app.Config.MakeBasicRequest("POST", sendURL, "", request)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer emailResp.Body.Close()

	checkResponseStatus(emailResp)

	scanner := bufio.NewScanner(os.Stdin)
	var code string

	fmt.Println(" < An email with a verification code has been sent to user's email. Please type in the code to continue > ")
	fmt.Println(" < Type 'exit' to cancel > ")
	for {
		fmt.Print(" > ")
		scanner.Scan()
		code = scanner.Text()
		if len(code) == 0 {
			fmt.Println(" < Please enter the verification code provided within the email below. > ")
			continue
		}

		if code == "exit" {
			return nil
		}

		if code == "resend" {
			emailRes, err := app.Config.MakeBasicRequest("POST", sendURL, "", request)
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

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("PUT", updateURL, token, updateReq)
	})
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	checkResponseStatus(resp)

	var updated models.UpdatedPassword

	if err = json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	params := database.UpdatePasswordParams{
		HashedPassword: updated.HashPassword,
		UpdatedAt: time.Now().Format("2006-01-02"),
		Name: creds.User.Name,
	}

	err = app.Config.Db.UpdatePassword(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error updating local user record: %w", err)
	}

	fmt.Println("Password updated successfully")
	return nil
}

//Resets a user's forgotten password, allowing for account recovery. Requires a verified email address.
func (app *CLIApp) commandResetPassword(args []string) error {

	email := args[0]

	sendURL := app.Config.Client.BaseURL + "/api/auth/email/send"
	resetURL := app.Config.Client.BaseURL + "/api/auth/reset-password"

	user, err := app.Config.Db.GetUserByEmail(context.Background(), email)
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

	emailResp, err := app.Config.MakeBasicRequest("POST", sendURL, "", request)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer emailResp.Body.Close()

	checkResponseStatus(emailResp)

	scanner := bufio.NewScanner(os.Stdin)
	var code string

	fmt.Println(" < An email with a verification code has been sent to user's email. Please type in the code to continue > ")
	fmt.Println(" < Type 'exit' to cancel > ")
	for {
		fmt.Print(" > ")
		scanner.Scan()
		code = scanner.Text()

		if len(code) == 0 {
			fmt.Println(" < Please enter the verification code provided within the email below. > ")
			continue
		}

		if code == "exit" {
			return nil
		}

		if code == "resend" {
			emailRes, err := app.Config.MakeBasicRequest("POST", sendURL, "", request)
			if err != nil {
				return fmt.Errorf("error making http request: %w", err)
			}
			defer emailRes.Body.Close()

			checkResponseStatus(emailRes)

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

	resetResp, err := app.Config.MakeBasicRequest("POST", resetURL, "", resetReq)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer resetResp.Body.Close()

	checkResponseStatus(resetResp)

	var updated models.UpdatedPassword

	if err = json.NewDecoder(resetResp.Body).Decode(&updated); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	params := database.UpdatePasswordParams{
		HashedPassword: updated.HashPassword,
		UpdatedAt: time.Now().Format("2006-01-02"),
		Name: user.Name,
	}

	err = app.Config.Db.UpdatePassword(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error updating local user record: %w", err)
	}

	fmt.Println("Password successfully reset")
	return nil
}

//Simple bufio scanner for getting an item nickname
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