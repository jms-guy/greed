package charts

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/guptarohit/asciigraph"
	"github.com/jms-guy/greed/models"
)

// Format a visual income vs expenses graph for data
func MakeIncomeChart(data []models.MonetaryData) {
	income := []float64{}
	expenses := []float64{}

	for _, item := range data {
		incomeStr := strings.TrimPrefix(item.Income, "-")
		i, _ := strconv.ParseFloat(incomeStr, 64)
		income = append(income, i)

		expensesStr := strings.TrimPrefix(item.Expenses, "-")
		e, _ := strconv.ParseFloat(expensesStr, 64)
		expenses = append(expenses, e)
	}

	slices.Reverse(income)
	slices.Reverse(expenses)

	graph := asciigraph.PlotMany(
		[][]float64{income, expenses},
		asciigraph.SeriesColors(asciigraph.Blue, asciigraph.Red),
		asciigraph.SeriesLegends("Income", "Expenses"),
		asciigraph.Caption("Income vs. Expenses - 24 month history"),
		asciigraph.Height(50),
		asciigraph.LowerBound(0),
		asciigraph.Width(210))

	fmt.Println(graph)
}
