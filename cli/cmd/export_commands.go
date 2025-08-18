package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/cli/internal/utils"
	"github.com/spf13/cobra"
)

// Function gets the export directory determined by operating system, retrieves transaction records from local database,
// and creates an exported .csv file containing those records
func (app *CLIApp) commandExportData(cmd *cobra.Command, args []string) error {
	accountName := args[0]

	// Trim quotes included in argument by shell
	accountName = strings.TrimPrefix(accountName, "'")
	accountName = strings.TrimSuffix(accountName, "'")
	accountName = strings.TrimPrefix(accountName, "\"")
	accountName = strings.TrimSuffix(accountName, "\"")
	// Input sanitization
	accountName = strings.ReplaceAll(accountName, "/", "")
	accountName = strings.ReplaceAll(accountName, "\\", "")
	accountName = strings.ReplaceAll(accountName, "..", "")

	exportDirectory := app.getExportDirectory()

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting credentials")
		return nil
	}

	params := database.GetAccountParams{
		Name:   accountName,
		UserID: creds.User.ID.String(),
	}
	account, err := app.Config.Db.GetAccount(context.Background(), params)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error getting local record: %w", err), "Local database error")
		return nil
	}

	txns, err := app.Config.Db.GetTransactions(context.Background(), account.ID)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error getting local records: %w", err), "Local database error")
		return nil
	}

	if len(txns) == 0 {
		fmt.Printf("No transaction records found for account %s\n", accountName)
		return nil
	}

	filename := fmt.Sprintf("%s.csv", accountName)
	exportFile := filepath.Join(exportDirectory, filename)

	err = os.MkdirAll(exportDirectory, 0o750)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making directory: %w", err), "File error")
		return nil
	}

	// #nosec G304 - file variables are controlled, no user input
	file, err := os.Create(exportFile)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error creating export file: %w", err), "File error")
		return nil
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Amount", "CurrencyCode", "Date", "Merchant", "Payment Channel", "Category"}
	err = writer.Write(headers)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error writing csv headers: %w", err), "File error")
		return nil
	}

	for _, txn := range txns {

		amount := strconv.FormatFloat(txn.Amount, 'f', 2, 64)
		toWrite := []string{amount, txn.IsoCurrencyCode.String, txn.Date.String, txn.MerchantName.String, txn.PaymentChannel, txn.PersonalFinanceCategory}

		err = writer.Write(toWrite)
		if err != nil {
			LogError(app.Config.Db, cmd, fmt.Errorf("error writing csv line: %w", err), "File error")
			return nil
		}
	}

	fmt.Printf("Data successfully exported to %s\n", exportFile)
	return nil
}

// Gets the base export directory to send exported .csv files to. Directory is based on operating system
func (app *CLIApp) getExportDirectory() string {
	var baseDir string

	if app.Config.OperatingSystem == "linux" && utils.IsWSL() {
		baseDir = filepath.Join(os.Getenv("HOME"), ".config", "greed", "exports")
		return baseDir
	}

	if app.Config.OperatingSystem == "windows" {
		baseDir = filepath.Join(os.Getenv("USERPROFILE"), "Documents", "greed_exports")
	} else {
		baseDir = filepath.Join(os.Getenv("HOME"), ".config", "greed", "exports")
	}

	return baseDir
}
