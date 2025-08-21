package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/spf13/cobra"
)

// Logs an error message in client database, to be accessible for viewing by user later
func LogError(db *database.Queries, cmd *cobra.Command, err error, userMsg string) {
	fmt.Println(userMsg)

	timestamp := time.Now().UTC().Truncate(time.Second)
	command := cmd.Use
	errMsg := err.Error()

	params := database.CreateErrorLogParams{
		Timestamp:    timestamp.Format(time.RFC3339),
		Command:      command,
		ErrorMessage: errMsg,
	}

	dbErr := db.CreateErrorLog(context.Background(), params)

	if dbErr != nil {
		fmt.Println("Local database encountered error while logging error")
	}
}

// Function prints the latest 5 commands encounted on machine and logged in database
func (app *CLIApp) commandReadLogs(cmd *cobra.Command) error {
	logs, err := app.Config.Db.GetLogs(context.Background())
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error getting log records: %w", err), "Local database error")
		return err
	}

	if len(logs) == 0 {
		fmt.Println("No error logs found.")
		return nil
	}

	fmt.Println("Last 5 errors logged by machine:")
	for _, log := range logs {
		fmt.Printf("Time: %s Command: %s Message: %s\n", log.Timestamp, log.Command, log.ErrorMessage)
	}

	return nil
}
