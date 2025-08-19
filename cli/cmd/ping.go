package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// Sends a simple ping to the server health endpoint, for checking server connection
func (app *CLIApp) commandPing(cmd *cobra.Command) error {
	healthURL := app.Config.Client.BaseURL + "/api/health"

	// #nosec G107 - BaseURL non-configurable.
	resp, err := http.Get(healthURL)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("Server is healthy (HTTP 200 OK)")
	} else {
		fmt.Printf("Server responded with status: %s\n", resp.Status)
	}
	return nil
}
