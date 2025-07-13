package utils

import (
	"github.com/jms-guy/greed/cli/internal/charts"
	"github.com/jms-guy/greed/cli/internal/tables"
	"github.com/jms-guy/greed/models"
)

//Display the correct format for the information given
func Display(accountName, mode string, data []models.MonetaryData) error {
	switch mode {
	case "graph": 
		charts.MakeIncomeChart(data)
		return nil 
	default:
		tbl, err := tables.MakeTableForMonetaryAggregate(data, accountName)
		if err != nil {
			return err
		}
		tbl.Print()
		return nil
	}
}