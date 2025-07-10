package tables

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/jms-guy/greed/models"
)

//Takes slice of transaction records, and paginates the results into a table, displaying a base number of 20 transaction records at a time.
//Listens for pgUp/pgDown key presses to view through record pages
func PaginateTransactionsTable(txns []models.Transaction, accountName string) error {

	if len(txns) == 0 {
		return fmt.Errorf("no results to display")
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("error creating terminal screen: %w", err)
	}
	defer screen.Fini()

	err = screen.Init()
	if err != nil {
		return fmt.Errorf("error initializing terminal screen: %w", err)
	}

	pageSize := 20
	currentPage := 1 
	endIndex := 0

	for {
		screen.Clear()

		startIndex := (currentPage - 1) * pageSize
		var displayItems []models.Transaction 

		if startIndex >= len(txns) {

			displayItems = []models.Transaction{} 

		} else {

			endIndex = min(startIndex + pageSize, len(txns))
			displayItems = txns[startIndex:endIndex]
		}

		tbl := MakeTableForTransactions(displayItems, accountName) 
		tbl.Print()

		event := screen.PollEvent()

		switch event := event.(type) {
		case *tcell.EventKey:
			switch event.Key() {
			case tcell.KeyPgDn:
				if endIndex < len(txns) {
					currentPage++
					continue
				}
			case tcell.KeyPgUp:
				if currentPage > 1 {
					currentPage--
					continue
				}
			case tcell.KeyEscape:
				os.Exit(0)
			}
		}

		screen.Show()
	}
}