package cmd

import (
	"context"
	"fmt"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/config"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/models"
	"github.com/spf13/cobra"
)

// Clears default item/accounts set in settings file
func (app *CLIApp) commandClearDefaults(cmd *cobra.Command) error {
	clearItem := models.ItemName{}
	clearAcc := database.Account{}

	app.Config.Settings.DefaultItem = clearItem
	app.Config.Settings.DefaultAccount = clearAcc

	err := config.SaveSettingsToFile(app.Config.SettingsFP, app.Config.Settings)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error saving settings")
		return err
	}

	fmt.Println("Default settings for item/account cleared")

	return nil
}

// Sets a default account in the settings file, to be used in place of specific command arguments, to make user experience better
func (app *CLIApp) commandSetDefaultAccount(cmd *cobra.Command, args []string) error {
	accountName := args[0]

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting credentials")
		return err
	}

	params := database.GetAccountParams{
		Name:   accountName,
		UserID: creds.User.ID.String(),
	}
	account, err := app.Config.Db.GetAccount(context.Background(), params)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error getting local account: %w", err), "Local database error")
		return err
	}

	app.Config.Settings.DefaultAccount = account

	err = config.SaveSettingsToFile(app.Config.SettingsFP, app.Config.Settings)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error saving settings")
		return err
	}

	fmt.Printf("Default account set in settings: %s\n", accountName)

	return nil
}

// Sets a default item in the settings file, to be used in place of specific command arguments, to make user experience better
func (app *CLIApp) commandSetDefaultItem(cmd *cobra.Command, args []string) error {
	itemName := args[0]

	item, err := getItemFromServer(app, itemName)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting item")
		return err
	}

	app.Config.Settings.DefaultItem = item

	err = config.SaveSettingsToFile(app.Config.SettingsFP, app.Config.Settings)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error saving settings")
		return err
	}

	fmt.Printf("Default item set in settings: %s\n", itemName)

	return nil
}
