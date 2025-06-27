package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/config"
	"github.com/jms-guy/greed/models"
)

//Wrapper function for refreshing JWT token logic
func DoWithAutoRefresh(c *config.Config, doRequest func(string) (*http.Response, error)) (*http.Response, error) {
	creds, _ := auth.GetCreds(c.ConfigFP)
	resp, err := doRequest(creds.AccessToken)
	if err != nil {
		return nil, err 
	}

	if resp.StatusCode == 401 {
		resp.Body.Close()
		if err := refreshCreds(c); err != nil {
			return nil, err 
		}
		creds, _  = auth.GetCreds(c.ConfigFP)
		resp, err = doRequest(creds.AccessToken)
	}
	return resp, err
}

//Refreshs JWT and refresh token for user
func refreshCreds(c *config.Config) error {
	refreshURL := c.Client.BaseURL + "/api/auth/refresh"

	creds, err := auth.GetCreds(c.ConfigFP)
	if err != nil {
		return err
	}

	request := models.RefreshRequest{
		RefreshToken: creds.RefreshToken,
	}

	resp, err := c.MakeBasicRequest("POST", refreshURL, "", request)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		fmt.Println("Server error")
		return nil
	}

	if resp.StatusCode >= 400 {
		fmt.Printf("Bad request - %s", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error details: %s\n", string(body))
		return nil
	}

	var response models.RefreshResponse

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("error decoding response data: %w", err)
	}

	creds.AccessToken = response.AccessToken
	creds.RefreshToken = response.RefreshToken

	err = auth.StoreTokens(creds, c.ConfigFP)
	if err != nil {
		return fmt.Errorf("error storing new tokens: %w", err)
	}

	return nil
} 