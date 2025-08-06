package tables

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/jms-guy/greed/models"
	"github.com/rodaine/table"
)

// Make monetary data aggregate table
func MakeTableForMonetaryAggregate(data []models.MonetaryData, accountName string) (table.Table, error) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New(
		"   |Account",
		"     |     ",
		"   Date   ",
		"     |     ",
		"   Income   ",
		"     |     ",
		"   Expenses   ",
		"     |     ",
		"   Net Income   ",
	)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, m := range data {
		incomeStr := strings.TrimPrefix(m.Income, "-")
		income, err := strconv.ParseFloat(incomeStr, 64)
		if err != nil {
			return tbl, fmt.Errorf("error parsing float value: %w", err)
		}
		expensesStr := strings.TrimPrefix(m.Expenses, "-")
		expenses, err := strconv.ParseFloat(expensesStr, 64)
		if err != nil {
			return tbl, fmt.Errorf("error parsing float value: %w", err)
		}

		calculatedNetIncome := income - expenses
		formattedNetIncome := fmt.Sprintf("%.2f", calculatedNetIncome)

		tbl.AddRow(
			fmt.Sprintf("   |%s   ", accountName),
			"     |     ",
			fmt.Sprintf("   %s   ", m.Date),
			"     |     ",
			fmt.Sprintf("   %v   ", income),
			"     |     ",
			fmt.Sprintf("   %v   ", expenses),
			"     |     ",
			fmt.Sprintf("   %s   ", formattedNetIncome),
		)
	}

	return tbl, nil
}

// Make transaction summary table
func MakeTableForSummaries(summaries []models.MerchantSummary, accountName string) table.Table {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New(
		"|Account",
		"  |  ",
		"Month",
		"  |  ",
		"Merchant",
		"  |  ",
		"Transaction Count",
		"  |  ",
		"Total Amount",
		"  |  ",
		"Category",
	)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, s := range summaries {
		tbl.AddRow(
			fmt.Sprintf("|%s", accountName),
			"  |  ",
			s.Month,
			"  |  ",
			s.Merchant,
			"  |  ",
			s.TxnCount,
			"  |  ",
			s.TotalAmount,
			"  |  ",
			s.Category,
		)
	}

	return tbl
}
