package cmd

import (
	"log"

	"github.com/jms-guy/greed/cli/internal/config"
	"github.com/spf13/cobra"
)

// App struct holding CLI config, as well as all command methods
type CLIApp struct {
	Config *config.Config
}

// Initializes a new app struct
func NewCLIApp() *CLIApp {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
	}
	app := CLIApp{
		Config: cfg,
	}

	return &app
}

// Initializes cobra commands
func (app *CLIApp) RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "greed",
		Short: "Greed is a CLI tool for tracking user's financial data",
		Long:  "Greed is a CLI tool for tracking user's financial data, through accessing financial institutions",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	dCmd := app.deleteCmd()
	// dCmd.AddCommand(app.deleteUserCmd())
	dCmd.AddCommand(app.deleteItemCmd())

	gCmd := app.getCmd()
	gCmd.AddCommand(app.getAccountsCmd())
	gCmd.AddCommand(app.getTransactionsCmd())
	gCmd.AddCommand(app.getIncomeDataCmd())

	rootCmd.AddCommand(dCmd)
	rootCmd.AddCommand(gCmd)
	rootCmd.AddCommand(app.pingCmd())
	rootCmd.AddCommand(app.registerCmd())
	rootCmd.AddCommand(app.loginCmd())
	rootCmd.AddCommand(app.logoutCmd())
	rootCmd.AddCommand(app.verifyCmd())
	rootCmd.AddCommand(app.itemsCmd())
	rootCmd.AddCommand(app.changepwCmd())
	rootCmd.AddCommand(app.resetpwCmd())
	rootCmd.AddCommand(app.fetchCmd())
	rootCmd.AddCommand(app.syncCmd())
	rootCmd.AddCommand(app.updateCmd())
	rootCmd.AddCommand(app.renameCmd())
	rootCmd.AddCommand(app.infoCmd())
	rootCmd.AddCommand(app.exportDataCmd())
	rootCmd.AddCommand(app.addItemCmd())
	rootCmd.AddCommand(app.logsCmd())

	return rootCmd
}

// Executes commands
func Execute() {
	app := NewCLIApp()
	//Ugly err handling, LogError function is used inside of command functions
	if err := app.RootCmd().Execute(); err != nil {
		return
	}
}
