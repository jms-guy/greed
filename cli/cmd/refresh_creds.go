package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/models"
	"github.com/spf13/cobra"
)

// Wrapper function for refreshing JWT token logic
func DoWithAutoRefresh(app *CLIApp, doRequest func(string) (*http.Response, error)) (*http.Response, error) {
	creds, _ := auth.GetCreds(app.Config.ConfigFP)
	resp, err := doRequest(creds.AccessToken)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		_ = resp.Body.Close()
		if refreshErr := refreshCreds(app); refreshErr != nil {
			return nil, refreshErr
		}
		creds, _ = auth.GetCreds(app.Config.ConfigFP)
		resp, err = doRequest(creds.AccessToken)
	}
	return resp, err
}

// Refreshs JWT and refresh token for user - logs user out automatically if session is expired
func refreshCreds(app *CLIApp) error {
	refreshURL := app.Config.Client.BaseURL + "/api/auth/refresh"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		return err
	}

	request := models.RefreshRequest{
		RefreshToken: creds.RefreshToken,
	}

	resp, err := app.Config.MakeBasicRequest("POST", refreshURL, "", request)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		var errResp struct {
			Error string `json:"error"`
		}
		if err = json.Unmarshal(body, &errResp); err == nil {
			if errResp.Error == "Token is expired" {
				fmt.Println(" < User's session is expired, please re-login. > ")
				if err = app.commandUserLogout(&cobra.Command{Use: "auto-logout"}); err != nil {
					return fmt.Errorf("error logging user out: %w", err)
				}
			}
		}
		return nil
	}

	var response models.RefreshResponse

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	creds.AccessToken = response.AccessToken
	creds.RefreshToken = response.RefreshToken

	err = auth.StoreTokens(creds, app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error storing new tokens: %w", err)
	}

	return nil
}
