package utils

import (
	"github.com/jms-guy/greed/cli/internal/charts"
	"github.com/jms-guy/greed/cli/internal/tables"
	"github.com/jms-guy/greed/models"
)

func Display(accountName, mode string, data []models.MonetaryData) error {
	switch mode {
	case "chart": 
		err := charts.MakeIncomeChart(data)
		if err != nil {
			return err 
		}
		return nil 
	case "graph":

		return nil 
	default:
		tbl := tables.MakeTableForMonetaryAggregate(data, accountName)
		tbl.Print()
		return nil
	}
}