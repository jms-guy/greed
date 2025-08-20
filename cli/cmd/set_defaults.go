package cmd

import "github.com/spf13/cobra"

func (app *CLIApp) commandSetDefaultAccount(cmd *cobra.Command, args []string) error {

}

func (app *CLIApp) commandSetDefaultItem(cmd *cobra.Command, args []string) error {
	itemName := args[0]

	item, err := getItemFromServer(app, itemName)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting item")
		return err
	}

}
