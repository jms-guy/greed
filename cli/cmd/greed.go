package cmd

import (
	"math"
	"github.com/spf13/cobra"
)

func (app *CLIApp) registerCmd() *cobra.Command {
    return &cobra.Command{
        Use:   		"register <name>",
		Aliases: 	[]string{"Register", "REGISTER"},
        Short: 		"Register a new user",
        Args: 		 cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            return app.commandRegisterUser(args)
        },
    }
}

func (app *CLIApp) loginCmd() *cobra.Command {
	return &cobra.Command{
		Use: 		"login <name>",
		Aliases: 	[]string{"Login", "LOGIN"},
		Short: 		"Login as a user",
		Args:		 cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.commandUserLogin(args)
		},
	}
}

func (app *CLIApp) logoutCmd() *cobra.Command {
	return &cobra.Command{
		Use: 		"logout",
		Aliases: 	[]string{"Logout", "LOGOUT"},
		Short: 		"Logs out of user credentials",
		Args: 		cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.commandUserLogout()
		},
	}
}

func (app *CLIApp) deleteCmd() *cobra.Command {
	return &cobra.Command{
		Use: 		"delete",
		Aliases: 	[]string{"Delete", "DELETE"},
		Short:  	"Deletes resources",
		Long:    	"Deletes users or items from the database and server",
	}
}

func (app *CLIApp) deleteUserCmd() *cobra.Command {
	return &cobra.Command{
		Use:   	 	"user <username>",
		Short: 		"Delete a user",
		Args:  		cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.commandDeleteUser(args)
		},
	}
}

func (app *CLIApp) deleteItemCmd() *cobra.Command {
	return &cobra.Command{
		Use:   		"item <item-name>",
		Short: 		"Delete an item",
		Args:  		cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.commandDeleteItem(args)
		},
	}
}

func (app *CLIApp) verifyCmd() *cobra.Command {
	return &cobra.Command{
		Use: 		"verify",
		Aliases: 	[]string{"Verify", "VERIFY"},
		Short: 		"Verifies a user's email address",
		Long: 		"Sends a verification code to user's email address",
		Args: 		cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.commandVerifyEmail()
		},
	}
}

func (app *CLIApp) itemsCmd() *cobra.Command {
	return &cobra.Command{
		Use: 		"items",
		Aliases: 	[]string{"Items", "ITEMS"},
		Short: 		"Lists a user's item records",
		Long: 		"Lists a user's item records. Items are financial institution connections, with each institution being one item",
		Args: 		cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.commandUserItems()
		},
	}
}

func (app *CLIApp) updatepwCmd() *cobra.Command {
	return &cobra.Command{
		Use: 		"updatepw",
		Aliases: 	[]string{"Updatepw", "UPDATEPW"},
		Short: 		"Updates a user's password",
		Args: 		cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.commandUpdatePassword()
		},
	}
}

func (app *CLIApp) resetpwCmd() *cobra.Command {
	return &cobra.Command{
		Use: 		"resetpw <email>",
		Aliases: 	[]string{"Resetpw", "RESETPW"},
		Short: 		"Resets a user's forgotten password",
		Args: 		cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.commandResetPassword(args)
		},
	}
}

func (app *CLIApp) fetchCmd() *cobra.Command {
	return &cobra.Command{
		Use: 		"fetch <item-name>",
		Aliases: 	[]string{"Fetch", "FETCH"},
		Short: 		"Fetchs account and transaction data for an item",
		Long: 		"Retrieves all account and transaction data for item from third party, populating database. Should only be used on a new item, afterwards use sync command",
		Args: 		cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := app.commandGetAccounts(args)
			if err != nil {
				return err 
			}
			return app.commandGetTransactions(args)
		},
	}
}

func (app *CLIApp) renameCmd() *cobra.Command {
	return &cobra.Command{
		Use: 		"rename <current-item-name> <new-item-name>",
		Aliases: 	[]string{"Rename", "RENAME"},
		Short: 		"Rename an item",
		Args: 		cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.commandRenameItem(args)
		},
	}
}

func (app *CLIApp) infoCmd() *cobra.Command {
	return &cobra.Command{
		Use: 		"info <account-name>",
		Aliases: 	[]string{"Info", "INFO"},
		Short: 		"Lists extended information for a given account",
		Args: 		cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.commandAccountInfo(args)
		},
	}
}

func (app *CLIApp) getCmd() *cobra.Command {
	return &cobra.Command{
		Use: 		"get",
		Aliases: 	[]string{"Get", "GET"},
		Short: 		"Lists resources",
		Long: 		"Returns a list of accounts or transactions",
	}
}

func (app *CLIApp) getAccountsCmd() *cobra.Command {
	return &cobra.Command{
		Use: 		"accounts [item-name]",
		Aliases: 	[]string{"Accounts", "ACCOUNTS"},
		Short: 		"Returns a list of accounts",
		Long: 		"Returns a list of accounts. If an item name is specified, it will return accounts only for that item. Otherwise it will return all accounts for user",
		Args: 		cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return app.commandListAccounts(args)
			} else {
				return app.commandListAllAccounts()
			}
		},
	}
}

func (app *CLIApp) getTransactionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: 		"transactions <account-name>",
		Aliases: 	[]string{"Transactions", "TRANSACTIONS", "txns", "Txns", "TXNS"},
		Short: 		"Returns a list of transactions for a given account",
		Long: 		"Returns transactions for an account, takes many optional flags that are used to build a query string",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accountName := args[0]
			merchant, _ := cmd.Flags().GetString("merchant")
			category, _ := cmd.Flags().GetString("category")
			channel, _ := cmd.Flags().GetString("channel")
			date, _ := cmd.Flags().GetString("date")
			start, _ := cmd.Flags().GetString("start")
			end, _ := cmd.Flags().GetString("end")
			min, _ := cmd.Flags().GetInt("min")
			max, _ := cmd.Flags().GetInt("max")
			limit, _ := cmd.Flags().GetInt("limit")
			order, _ := cmd.Flags().GetString("order")
			summary, _ := cmd.Flags().GetBool("summary")

			return app.commandGetTxnsAccount(accountName, merchant, category, channel, date, start, end, order, min, max, limit, summary)
		},
	}

	cmd.Flags().String("merchant", "", "Filter transactions by merchant name")
	cmd.Flags().String("category", "", "Filter transactions by category")
	cmd.Flags().String("channel", "", "Filter transactions by payment channel")
	cmd.Flags().String("date", "", "Filter transactions by date (format year-month-day {2006-01-02})")
	cmd.Flags().String("start", "", "Filters transactions by adding a starting date (format year-month-day {2006-01-02})")
	cmd.Flags().String("end", "", "Filters transactions by adding an ending date (format year-month-day {2006-01-02})")
	cmd.Flags().Int("min", math.MinInt64, "Filters transactions by a minimum amount (negative means income)")
	cmd.Flags().Int("max", math.MaxInt64, "Filters transactions by a maximum amount")
	cmd.Flags().Int("limit", 100, "Filters transactions by limiting the number shown")
	cmd.Flags().String("order", "", "Filters transactions by re-ordering by date [ASC | DESC]")
	cmd.Flags().Bool("summary", false, "Provides a summary of transactions. Overrides most other flags. Useful with the [date] flag")

	return cmd
}

func (app *CLIApp) getIncomeData() *cobra.Command {
	cmd := &cobra.Command{
		Use: "income",
		Aliases: []string{"Income", "INCOME", "inc", "INC"},
		Short: "Returns aggregate income/expenses data for account history",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accountName := args[0]
			mode, _ := cmd.Flags().GetString("mode")

			return app.commandGetIncome(accountName, mode)
		},
	}

	cmd.Flags().String("mode", "table" , "Change visual output of data [graph | chart]")

	return cmd
}