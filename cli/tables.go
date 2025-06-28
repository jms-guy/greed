package main

import (

	"github.com/fatih/color"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/models"
	"github.com/rodaine/table"
)

//Format a table for transaction records
func MakeTableForTransactions(txns []models.Transaction, accountName string) table.Table {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New(
		"Account",
		"Amount",
		"Date",
		"Merchant Name",
		"Payment Channel",
		"Category",
		"Currency Code",
	)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, txn := range txns {
		tbl.AddRow(
			accountName,
			txn.Amount,
			txn.Date.Format("2006-01-02"),
			txn.MerchantName,
			txn.PaymentChannel,
			txn.PersonalFinanceCategory,
			txn.IsoCurrencyCode,
		)
	}

	return tbl
}

//Format a table for a single account record
func MakeSingleAccountTable(acc database.Account) table.Table {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New(
		"Institution",
	 	"ID", 
	 	"Name",
	  	"Type",
	  	"Subtype",
		"Mask",
		"Official Name",
	    "Available Bal.",
		"Current Bal.",
		"Currency Code",
		"Created at",
		"Updated at",
	)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	tbl.AddRow(
	acc.InstitutionName.String,
	acc.ID,
	acc.Name,
	acc.Type,
	acc.Subtype.String,
	acc.Mask.String,
	acc.OfficialName.String,
	acc.AvailableBalance.Float64,
	acc.CurrentBalance.Float64,
	acc.IsoCurrencyCode.String,
	acc.CreatedAt,
	acc.UpdatedAt,
	)

	return tbl
}

//Format a table of accounts data to display of single item record
func MakeAccountsTable(accounts []models.Account, institutionName string) table.Table {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Institution", "Name", "Type", "Subtype", "Available Bal.", "Current Bal.", "Currency Code")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, acc := range accounts {
		tbl.AddRow( 
		institutionName,
		acc.Name,
		acc.Type,
		acc.Subtype,
		acc.AvailableBalance,
		acc.CurrentBalance,
		acc.IsoCurrencyCode)
	}

	return tbl
}

//Make table of accounts data to display of all user items
func MakeAccountsTableAllItems(accounts []database.Account) table.Table {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Institution", "Name", "Type", "Subtype", "Available Bal.", "Current Bal.", "Currency Code")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, acc := range accounts {
		tbl.AddRow( 
		acc.InstitutionName.String,
		acc.Name,
		acc.Type,
		acc.Subtype.String,
		acc.AvailableBalance.Float64,
		acc.CurrentBalance.Float64,
		acc.IsoCurrencyCode.String)
	}

	return tbl
}
