package cmd

import (
	"fmt"
	"strings"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/spf13/cobra"
)

// Syncs database with updated account balances, and transaction records
// Updates account balances, and deletes local transaction records before replacing them with
// updated records from server database.
func (app *CLIApp) commandSync(cmd *cobra.Command, args []string) error {
	itemName := args[0]
	itemsURL := app.Config.Client.BaseURL + "/api/items"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting credentials")
		return err
	}

	// Get item ID and institution
	itemID, itemInst, err := findItemHelper(app, itemName, itemsURL)
	if err != nil {
		if strings.Contains(err.Error(), "no item found") {
			LogError(app.Config.Db, cmd, err, "No item found")
			return err
		} else {
			LogError(app.Config.Db, cmd, err, "Error contacting server")
			return err
		}
	}

	err = syncAccountBalances(app, creds, itemID, itemInst)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error syncing account data")
		return err
	}

	fmt.Println(" > Fetching transaction data...")

	err = processWebhookRecords(app, itemID, "ITEM", "DEFAULT_UPDATE")
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error processing webhooks")
		return err
	}

	err = syncTransactionData(app, creds, itemID)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error syncing transaction data")
		return err
	}

	webhookCodes := []string{"TRANSACTIONS_UPDATES_AVAILABLE", "TRANSACTIONS_REMOVED", "DEFAULT_UPDATE", "INITIAL_UPDATE", "HISTORICAL_UPDATE", "SYNC_UPDATES_AVAILABLE"}

	for _, code := range webhookCodes {
		err = processWebhookRecords(app, itemID, code, "TRANSACTIONS")
		if err != nil {
			LogError(app.Config.Db, cmd, err, "Error contacting server")
			return err
		}
	}

	return nil
}
