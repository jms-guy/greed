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
		tbl, err := tables.MakeTableForMonetaryAggregate(data, accountName)
		if err != nil {
			return err
		}
		tbl.Print()
		return nil
	}
}