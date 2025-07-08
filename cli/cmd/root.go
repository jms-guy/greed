package cmd

import (
	"fmt"
	"log"
	"os"
	"github.com/jms-guy/greed/cli/internal/config"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

//App struct holding CLI config, as well as all command methods
type CLIApp struct {
    Config *config.Config
}

//Initializes a new app struct
func NewCLIApp() *CLIApp {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
	}
	app := CLIApp{
		Config: cfg,
	}

	return &app
}

//Initializes cobra commands
func (app *CLIApp) RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:		"greed",
		Short: 		"Greed is a CLI tool for tracking user's financial data",
		Long: 		"Greed is a CLI tool for tracking user's financial data, through accessing financial institutions",
		Run: 		func(cmd *cobra.Command, args[]string) {
	
		},
	}
	dCmd := app.deleteCmd()
	dCmd.AddCommand(app.deleteUserCmd())
	dCmd.AddCommand(app.deleteItemCmd())

	gCmd := app.getCmd()
	gCmd.AddCommand(app.getAccountsCmd())
	gCmd.AddCommand(app.getTransactionsCmd())
	gCmd.AddCommand(app.getIncomeData())

	rootCmd.AddCommand(dCmd)
	rootCmd.AddCommand(gCmd)
	rootCmd.AddCommand(app.pingCmd())
	rootCmd.AddCommand(app.registerCmd())
	rootCmd.AddCommand(app.loginCmd())
	rootCmd.AddCommand(app.logoutCmd())
	rootCmd.AddCommand(app.verifyCmd())
	rootCmd.AddCommand(app.itemsCmd())
	rootCmd.AddCommand(app.updatepwCmd())
	rootCmd.AddCommand(app.resetpwCmd())
	rootCmd.AddCommand(app.fetchCmd())
	rootCmd.AddCommand(app.syncCmd())
	rootCmd.AddCommand(app.renameCmd())
	rootCmd.AddCommand(app.infoCmd())
	rootCmd.AddCommand(app.exportData())

	return rootCmd
}

//Executes commands
func Execute() {
	app := NewCLIApp()

	if err := app.RootCmd().Execute(); err != nil {
		fmt.Printf("An error occurred while executing: %s\n", err)
		os.Exit(1)
	}
}