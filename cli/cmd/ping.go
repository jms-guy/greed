package cmd

import (
	"fmt"
	"net/http"
)

// Sends a simple ping to the server health endpoint, for checking server connection
func (app *CLIApp) commandPing() error {
	healthURL := app.Config.Client.BaseURL + "/api/health"

	// #nosec G107 - BaseURL is user-configurable for user-hosted backends; scheme (http/https) is validated in LoadConfig.
	resp, err := http.Get(healthURL)
	if err != nil {
		fmt.Printf("Error sending health check: %s\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("Server is healthy (HTTP 200 OK)")
	} else {
		fmt.Printf("Server responded with status: %s\n", resp.Status)
	}
	return nil
}
