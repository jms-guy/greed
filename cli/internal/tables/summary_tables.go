package tables

import (

	"github.com/fatih/color"
	"github.com/jms-guy/greed/models"
	"github.com/rodaine/table"
)
//Make monetary data aggregate table
func MakeTableForMonetaryAggregate(data []models.MonetaryData, accountName string) table.Table {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New(
		"Account",
		"Date",
		"Income",
		"Expenses",
		"Net Income",
	)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, m := range data {
		tbl.AddRow(
			accountName,
			m.Date,
			m.Income,
			m.Expenses,
			m.NetIncome,
		)
	}

	return tbl
}

//Make transaction summary table
func MakeTableForSummaries(summaries []models.MerchantSummary, accountName string) table.Table {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New(
		"Account",
		"Month",
		"Merchant",
		"Transaction Count",
		"Total Amount",
		"Category",
	)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, s := range summaries {
		tbl.AddRow(
			accountName,
			s.Month,
			s.Merchant,
			s.TxnCount,
			s.TotalAmount,
			s.Category,
		)
	}

	return tbl
}