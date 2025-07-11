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
)

//Function gets the export directory determined by operating system, retrieves transaction records from local database,
//and creates an exported .csv file containing those records
func (app *CLIApp) commandExportData(args []string) error {
	accountName := args[0]

	//Trim quotes included in argument by shell
	accountName = strings.TrimPrefix(accountName, "'")
	accountName = strings.TrimSuffix(accountName, "'")
	accountName = strings.TrimPrefix(accountName, "\"")
	accountName = strings.TrimSuffix(accountName, "\"")

	exportDirectory := app.getExportDirectory()
	if len(args) == 2 {
		exportDirectory = args[1]
	}
	
	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	user, err := app.Config.Db.GetUser(context.Background(), creds.User.Name)
	if err != nil {
		return fmt.Errorf("error getting local user record: %w", err)
	}

	params := database.GetAccountParams{
		Name: accountName,
		UserID: user.ID,
	}
	account, err := app.Config.Db.GetAccount(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error getting local account record: %w", err)
	}

	txns, err := app.Config.Db.GetTransactions(context.Background(), account.ID)
	if err != nil {
		return fmt.Errorf("error getting transaction records: %w", err)
	}

	if len(txns) == 0 {
		fmt.Printf("No transaction records found for account %s\n", accountName)
		return nil 
	}

	filename :=fmt.Sprintf("%s.csv", accountName)
	exportFile := filepath.Join(exportDirectory, filename)

	err = os.MkdirAll(exportDirectory, 0755)
	if err != nil {
		return fmt.Errorf("error creating export directory %s: %w", exportDirectory, err)
	}

	file, err := os.Create(exportFile)
	if err != nil {
		return fmt.Errorf("error creating export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Amount", "CurrencyCode", "Date", "Merchant", "Payment Channel", "Category"}
	err = writer.Write(headers)
	if err != nil {
		return fmt.Errorf("error writing CSV headers: %w", err)
	}

	for _, txn := range txns {

		amount := strconv.FormatFloat(txn.Amount, 'f', 2, 64)
		toWrite := []string{amount, txn.IsoCurrencyCode.String, txn.Date.String, txn.MerchantName.String, txn.PaymentChannel, txn.PersonalFinanceCategory}

		err = writer.Write(toWrite)
		if err != nil {
			return fmt.Errorf("error writing CSV line: %w", err)
		}
	}

	fmt.Printf("Data successfully exported to %s\n", exportFile)
	return nil
}

//Gets the base export directory to send exported .csv files to. Directory is based on operating system
func (app *CLIApp) getExportDirectory() string {
	var baseDir string

	if app.Config.OperatingSystem == "linux" && utils.IsWSL() {
		baseDir = os.Getenv("CSV_EXPORT_BASE_DIR_LINUX_MAC")
		if baseDir == "" {
			baseDir = filepath.Join(os.Getenv("HOME"), "greed_exports")
		}
		return baseDir
	}

	if app.Config.OperatingSystem == "windows" {
		baseDir = os.Getenv("CSV_EXPORT_BASE_DIR_WINDOWS")
	} else {
		baseDir = os.Getenv("CSV_EXPORT_BASE_DIR_LINUX_MAC")
	}

	if baseDir == "" {
		if app.Config.OperatingSystem == "windows" {
			baseDir = filepath.Join(os.Getenv("USERPROFILE"), "Documents", "greed_exports")
		} else {
			baseDir = filepath.Join(os.Getenv("HOME"), "greed_exports")
		}
	}

	return baseDir
}