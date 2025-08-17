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
