package tables

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/jms-guy/greed/models"
)

//Takes slice of transaction records, and paginates the results into a table, displaying a base number of 20 transaction records at a time.
//Listens for pgUp/pgDown key presses to view through record pages
func PaginateTransactionsTable(txns []models.Transaction, accountName string, pageSize int) error {

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

	currentPage := 1 
	endIndex := 0

	for {
		screen.Clear()

		//Determine indexes of transaction items to display
		startIndex := (currentPage - 1) * pageSize
		var displayItems []models.Transaction 

		if startIndex >= len(txns) {

			displayItems = []models.Transaction{} 

		} else {

			endIndex = min(startIndex + pageSize, len(txns))
			displayItems = txns[startIndex:endIndex]
		}

		CreateTable(screen, displayItems, accountName)
		screen.Show()
		
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
				return nil
			}
		}
	}
}

//Draws a table of transaction data onto the tcell screen
func CreateTable(screen tcell.Screen, displayItems []models.Transaction, accountName string) {
	//Define tcell screen styles and variables to create table
	headerStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Underline(true)
	columnStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)

	columnHeaders := []string{"Account", "Amount", "Date", "Merchant Name", "Payment Channel", "Category", "Currency Code"}
	columnWidths := []int{10, 10, 12, 20, 15, 15, 10} 
	columnPadding := 3

	currentX, currentY := 0, 0

	//Draw table headers
	for i, header := range columnHeaders {
		for _, r := range header {
			screen.SetContent(currentX, currentY, r, nil, headerStyle)
			currentX++
		}

		currentX += columnWidths[i] - len(header) + columnPadding
	}
	currentY += 2

	//Draw transaction rows
	for _, txn := range displayItems {
		currentX = 0

		accountStr := accountName 
		if len(accountStr) > columnWidths[0] {
			accountStr = accountStr[:columnWidths[0]]
		}
		for _, r := range fmt.Sprintf("%-*s", columnWidths[0], accountStr) {
			screen.SetContent(currentX, currentY, r, nil, columnStyle)
			currentX++
		}
		currentX += columnPadding

		amountStr := txn.Amount
		if len(amountStr) > columnWidths[1] {
			amountStr = amountStr[:columnWidths[1]]
		}
		for _, r := range fmt.Sprintf("%-*s", columnWidths[1], amountStr) {
			screen.SetContent(currentX, currentY, r, nil, columnStyle)
			currentX++
		}
		currentX += columnPadding

		dateStr := txn.Date.Format("2006-01-02")
		if len(dateStr) > columnWidths[2] {
			dateStr = dateStr[:columnWidths[2]]
		}
		for _, r := range fmt.Sprintf("%-*s", columnWidths[2], dateStr) {
			screen.SetContent(currentX, currentY, r, nil, columnStyle)
			currentX++
		}
		currentX += columnPadding

		merchantStr := txn.MerchantName
		if len(merchantStr) > columnWidths[3] {
			merchantStr = merchantStr[:columnWidths[3]]
		}
		for _, r := range fmt.Sprintf("%-*s", columnWidths[3], merchantStr) {
			screen.SetContent(currentX, currentY, r, nil, columnStyle)
			currentX++
		}
		currentX += columnPadding

		channelStr := txn.PaymentChannel
		if len(channelStr) > columnWidths[4] {
			channelStr = channelStr[:columnWidths[4]]
		}
		for _, r := range fmt.Sprintf("%-*s", columnWidths[4], channelStr) {
			screen.SetContent(currentX, currentY, r , nil, columnStyle)
			currentX++
		}
		currentX += columnPadding

		categoryStr := txn.PersonalFinanceCategory
		if len(categoryStr) > columnWidths[5] {
			categoryStr = categoryStr[:columnWidths[5]]
		}
		for _, r := range fmt.Sprintf("%-*s", columnWidths[5], categoryStr) {
			screen.SetContent(currentX, currentY, r, nil, columnStyle)
			currentX++
		}
		currentX += columnPadding

		currencyStr := txn.IsoCurrencyCode
		if len(currencyStr) > columnWidths[6] {
			currencyStr = currencyStr[:columnWidths[6]]
		}
		for _, r := range fmt.Sprintf("%-*s", columnWidths[6], currencyStr) {
			screen.SetContent(currentX, currentY, r, nil, columnStyle)
			currentX++
		}

		currentY++
	}

	currentX = 0
	exitStr := "Press the 'esc' key to close table."
	instructions := "Use the 'pageUp' and 'pageDown' keys to scroll table."
	for _, r := range exitStr {
		screen.SetContent(currentX, currentY, r, nil, headerStyle)
		currentX++
	}
	currentY++
	currentX = 0
	for _, r := range instructions {
		screen.SetContent(currentX, currentY, r, nil, headerStyle)
		currentX++
	}
}

func PaginateSummariesTable(summaries []models.MerchantSummary, accountName string, pageSize int) error {

	if len(summaries) == 0 {
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

	currentPage := 1 
	endIndex := 0

	for {
		screen.Clear()

		//Determine indexes of transaction items to display
		startIndex := (currentPage - 1) * pageSize
		var displayItems []models.MerchantSummary 

		if startIndex >= len(summaries) {

			displayItems = []models.MerchantSummary{} 

		} else {

			endIndex = min(startIndex + pageSize, len(summaries))
			displayItems = summaries[startIndex:endIndex]
		}

		CreateSummariesTable(screen, displayItems, accountName)
		screen.Show()
		
		event := screen.PollEvent()

		switch event := event.(type) {
		case *tcell.EventKey:
			switch event.Key() {
			case tcell.KeyPgDn:
				if endIndex < len(summaries) {
					currentPage++
					continue
				}
			case tcell.KeyPgUp:
				if currentPage > 1 {
					currentPage--
					continue
				}
			case tcell.KeyEscape:
				return nil
			}
		}
	}
}

//Draws a table of transaction data onto the tcell screen
func CreateSummariesTable(screen tcell.Screen, displayItems []models.MerchantSummary, accountName string) {
	//Define tcell screen styles and variables to create table
	headerStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Underline(true)
	columnStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)

	columnHeaders := []string{"Month", "Merchant", "Txn. Count", "Category", "Total Amount"}
	columnWidths := []int{10, 10, 10, 15, 12} 
	columnPadding := 3

	currentX, currentY := 0, 0

	//Draw table headers
	for i, header := range columnHeaders {
		for _, r := range header {
			screen.SetContent(currentX, currentY, r, nil, headerStyle)
			currentX++
		}

		currentX += columnWidths[i] - len(header) + columnPadding
	}
	currentY += 2

	//Draw transaction rows
	for _, sum := range displayItems {
		currentX = 0

		monthStr := sum.Month
		if len(monthStr) > columnWidths[0] {
			monthStr = monthStr[:columnWidths[0]]
		}
		for _, r := range fmt.Sprintf("%-*s", columnWidths[0], monthStr) {
			screen.SetContent(currentX, currentY, r, nil, columnStyle)
			currentX++
		}
		currentX += columnPadding

		merchantStr := sum.Merchant
		if len(merchantStr) > columnWidths[1] {
			merchantStr = merchantStr[:columnWidths[1]]
		}
		for _, r := range fmt.Sprintf("%-*s", columnWidths[1], merchantStr) {
			screen.SetContent(currentX, currentY, r, nil, columnStyle)
			currentX++
		}
		currentX += columnPadding

		countStr := strconv.Itoa(int(sum.TxnCount))
		if len(countStr) > columnWidths[2] {
			countStr = countStr[:columnWidths[2]]
		}
		for _, r := range fmt.Sprintf("%-*s", columnWidths[2], countStr) {
			screen.SetContent(currentX, currentY, r, nil, columnStyle)
			currentX++
		}
		currentX += columnPadding

		categoryStr := sum.Category
		if len(categoryStr) > columnWidths[3] {
			categoryStr = categoryStr[:columnWidths[3]]
		}
		for _, r := range fmt.Sprintf("%-*s", columnWidths[3],categoryStr) {
			screen.SetContent(currentX, currentY, r, nil, columnStyle)
			currentX++
		}
		currentX += columnPadding

		amountStr := sum.TotalAmount
		if len(amountStr) > columnWidths[4] {
			amountStr = amountStr[:columnWidths[4]]
		}
		for _, r := range fmt.Sprintf("%-*s", columnWidths[4], amountStr) {
			screen.SetContent(currentX, currentY, r , nil, columnStyle)
			currentX++
		}

		currentY++
	}

	currentX = 0
	exitStr := "Press the 'esc' key to close table."
	instructions := "Use the 'pageUp' and 'pageDown' keys to scroll table."
	for _, r := range exitStr {
		screen.SetContent(currentX, currentY, r, nil, headerStyle)
		currentX++
	}
	currentY++
	currentX = 0
	for _, r := range instructions {
		screen.SetContent(currentX, currentY, r, nil, headerStyle)
		currentX++
	}
}