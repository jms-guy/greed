package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/models"
	"github.com/spf13/cobra"
)

// Function creates a new user record on server, and in loca database
func (app *CLIApp) commandRegisterUser(cmd *cobra.Command, args []string) error {
	username := args[0]

	registerURL := app.Config.Client.BaseURL + "/api/auth/register"
	sendURL := app.Config.Client.BaseURL + "/api/auth/email/send"
	verifyURL := app.Config.Client.BaseURL + "/api/auth/email/verify"

	// Get user password
	password, err := registerPasswordHelper()
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error encountered getting password")
		return err
	}

	// Get user email
	email, err := registerEmailHelper()
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error encountered registering email")
		return err
	}
	if email == "exit" {
		return nil
	}

	reqData := models.UserDetails{
		Name:     username,
		Password: password,
		Email:    email,
	}

	res, err := app.Config.MakeBasicRequest("POST", registerURL, "", reqData)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer res.Body.Close()

	err = checkResponseStatus(res)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	var user models.User

	if err = json.NewDecoder(res.Body).Decode(&user); err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
		return err
	}

	emailData := models.EmailVerification{
		UserID: user.ID,
		Email:  user.Email,
	}

	emailRes, err := app.Config.MakeBasicRequest("POST", sendURL, "", emailData)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer emailRes.Body.Close()

	err = checkResponseStatus(emailRes)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	// Email verification flow
	verified, err := verifyEmailHelper(app, sendURL, verifyURL, user, emailData)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error encountered verifying email")
		return err
	}
	if !verified {
		return nil
	}

	params := database.CreateUserParams{
		ID:             user.ID.String(),
		Name:           username,
		CreatedAt:      user.CreatedAt.Format("2006-01-02"),
		UpdatedAt:      user.UpdatedAt.Format("2006-01-02"),
		HashedPassword: user.HashedPassword,
		Email:          user.Email,
		IsVerified:     sql.NullBool{Bool: true, Valid: true},
	}

	_, err = app.Config.Db.CreateUser(context.Background(), params)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error creating user record: %w", err), "Local database error")
		return err
	}

	fmt.Printf("User: %s has been successfully registered!\n", username)
	fmt.Println("As a demo user, you have 10 total uses for commands (fetch, sync). The initial fetch will use 2, and each sync afterwards will also use 2.")

	return nil
}

// Function creates a login session for user, getting auth tokens
func (app *CLIApp) commandUserLogin(cmd *cobra.Command, args []string) error {
	username := args[0]

	loginURL := app.Config.Client.BaseURL + "/api/auth/login"
	itemsURL := app.Config.Client.BaseURL + "/api/items"

	// Get user credentials
	login, err := userLoginHelper(app, username, loginURL)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error logging in")
		return err
	}

	// Check user for existing items
	items, err := userCheckItemsHelper(app, login, itemsURL)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error logging in")
		return err
	}
	if len(items) != 0 {
		err = checkForWebhookRecords(app, items)
		if err != nil {
			LogError(app.Config.Db, cmd, err, "Error contacting server")
			return err
		}

		return nil
	}

	// If no items for user found, this is determined to be first time login
	// Go through first time Plaid Link flow
	linked, err := userFirstTimePlaidLinkHelper(app, login, itemsURL)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error linking financial institution")
		return err
	}
	if linked {
		return nil
	}

	err = auth.StoreTokens(login, app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error logging in")
		return err
	}

	return nil
}

// Logs a user out, by deleting their local credentials file, and expiring their session delegation server side
func (app *CLIApp) commandUserLogout(cmd *cobra.Command) error {
	logoutURL := app.Config.Client.BaseURL + "/api/auth/logout"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error reading file, could not log out")
		return err
	}

	request := models.RefreshRequest{
		RefreshToken: creds.RefreshToken,
	}

	resp, err := app.Config.MakeBasicRequest("POST", logoutURL, "", request)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	err = auth.RemoveCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error logging out")
		return err
	}

	fmt.Println("Logged out successfully!")
	return nil
}

/*
// Delete a user's records locally, and server side
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

	err = checkResponseStatus(resp)
	if err != nil {
		return err
	}

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
*/
// Verifies a user's email address
func (app *CLIApp) commandVerifyEmail(cmd *cobra.Command) error {
	sendURL := app.Config.Client.BaseURL + "/api/auth/email/send"
	verifyURL := app.Config.Client.BaseURL + "/api/auth/email/verify"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting credentials")
		return err
	}

	user, err := app.Config.Db.GetUser(context.Background(), creds.User.Name)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error getting user record: %w", err), "Local database error")
		return err
	}

	if user.IsVerified.Bool {
		fmt.Println("User already has a verified email address")
		return nil
	}

	sendReq := models.EmailVerification{
		UserID: creds.User.ID,
		Email:  user.Email,
	}

	sendResp, err := app.Config.MakeBasicRequest("POST", sendURL, "", sendReq)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer sendResp.Body.Close()

	err = checkResponseStatus(sendResp)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	code, err := getEmailCodeHelper(app, creds.User.Email, sendURL, sendReq)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}
	if code == "" {
		return nil
	}

	verifyData := models.EmailVerificationWithCode{
		UserID: creds.User.ID,
		Code:   code,
	}

	verifyRes, err := app.Config.MakeBasicRequest("POST", verifyURL, "", verifyData)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer verifyRes.Body.Close()

	err = checkResponseStatus(verifyRes)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	verifyParams := database.VerifyEmailParams{
		IsVerified: sql.NullBool{Bool: true, Valid: true},
		UpdatedAt:  time.Now().Format("2006-01-02"),
		Name:       user.Name,
	}

	err = app.Config.Db.VerifyEmail(context.Background(), verifyParams)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Local database error")
		return err
	}

	fmt.Printf("User '%s' email '%s' verified successfully\n", user.Name, user.Email)

	return nil
}

// Lists all items attached to a specific user
func (app *CLIApp) commandUserItems(cmd *cobra.Command) error {
	itemsURL := app.Config.Client.BaseURL + "/api/items"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting credentials")
		return err
	}

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", itemsURL, token, nil)
	})
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	if resp.StatusCode == 200 {
		var itemsResponse struct {
			Items []models.ItemName `json:"items"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&itemsResponse); err != nil {
			LogError(app.Config.Db, cmd, err, "Error contacting server")
			return err
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

// Updates a user's password in record. Requires verified email address to send code to
func (app *CLIApp) commandChangePassword(cmd *cobra.Command) error {
	sendURL := app.Config.Client.BaseURL + "/api/auth/email/send"
	updateURL := app.Config.Client.BaseURL + "/api/users/update-password"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting credentials")
		return err
	}

	pw, err := auth.ReadPassword("Please enter your current password > ")
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting password")
		return err
	}

	err = auth.ValidatePasswordHash(creds.User.HashedPassword, pw)
	if err != nil {
		fmt.Println("Bad password input")
		return err
	}

	password, err := registerPasswordHelper()
	if err != nil {
		return err
	}

	request := models.EmailVerification{
		UserID: creds.User.ID,
		Email:  creds.User.Email,
	}

	emailResp, err := app.Config.MakeBasicRequest("POST", sendURL, "", request)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer emailResp.Body.Close()

	err = checkResponseStatus(emailResp)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	code, err := getEmailCodeHelper(app, creds.User.Email, sendURL, request)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}
	if code == "" {
		return nil
	}

	updateReq := models.UpdatePassword{
		NewPassword: password,
		Code:        code,
	}

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("PUT", updateURL, token, updateReq)
	})
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	var updated models.UpdatedPassword

	if err = json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
		return err
	}

	params := database.UpdatePasswordParams{
		HashedPassword: updated.HashPassword,
		UpdatedAt:      time.Now().Format("2006-01-02"),
		Name:           creds.User.Name,
	}

	err = app.Config.Db.UpdatePassword(context.Background(), params)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Local database error")
		return err
	}

	fmt.Println("Password updated successfully")
	return nil
}

// Resets a user's forgotten password, allowing for account recovery. Requires a verified email address.
func (app *CLIApp) commandResetPassword(cmd *cobra.Command, args []string) error {
	email := args[0]

	sendURL := app.Config.Client.BaseURL + "/api/auth/email/send"
	resetURL := app.Config.Client.BaseURL + "/api/auth/reset-password"

	user, err := app.Config.Db.GetUserByEmail(context.Background(), email)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error getting user record: %w", err), "Local database error")
		return err
	}

	password, err := registerPasswordHelper()
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error registering password")
		return err
	}

	request := models.EmailVerification{
		UserID: uuid.MustParse(user.ID),
		Email:  user.Email,
	}

	emailResp, err := app.Config.MakeBasicRequest("POST", sendURL, "", request)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer emailResp.Body.Close()

	err = checkResponseStatus(emailResp)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	code, err := getEmailCodeHelper(app, email, sendURL, request)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}
	if code == "" {
		return nil
	}

	resetReq := models.ResetPassword{
		Email:       user.Email,
		Code:        code,
		NewPassword: password,
	}

	resetResp, err := app.Config.MakeBasicRequest("POST", resetURL, "", resetReq)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer resetResp.Body.Close()

	err = checkResponseStatus(resetResp)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	var updated models.UpdatedPassword

	if err = json.NewDecoder(resetResp.Body).Decode(&updated); err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
		return err
	}

	params := database.UpdatePasswordParams{
		HashedPassword: updated.HashPassword,
		UpdatedAt:      time.Now().Format("2006-01-02"),
		Name:           user.Name,
	}

	err = app.Config.Db.UpdatePassword(context.Background(), params)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error updating user record: %w", err), "Error contacting server")
		return err
	}

	fmt.Println("Password successfully reset")
	return nil
}
