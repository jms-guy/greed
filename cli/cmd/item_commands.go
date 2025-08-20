package cmd

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/database"
	"github.com/jms-guy/greed/models"
	"github.com/spf13/cobra"
)

// Fetches all transaction records for accounts attached to given item
func (app *CLIApp) commandGetTransactions(cmd *cobra.Command, args []string) error {
	itemName := args[0]

	item, err := getItemFromServer(app, itemName)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting item")
		return err
	}

	txnsURL := app.Config.Client.BaseURL + "/api/items/" + item.ItemId + "/access/transactions"

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("POST", txnsURL, token, nil)
	})
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http req: %w", err), "Error contacting server")
		return err
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	var txns []models.Transaction
	if err = json.NewDecoder(resp.Body).Decode(&txns); err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
		return err
	}

	fmt.Println("Creating local records...")

	for _, t := range txns {
		a, err := strconv.ParseFloat(t.Amount, 64)
		if err != nil {
			LogError(app.Config.Db, cmd, fmt.Errorf("error converting string value: %w", err), "Data error")
			return err
		}

		params := database.CreateTransactionParams{
			ID:                      t.Id,
			AccountID:               t.AccountId,
			Amount:                  a,
			IsoCurrencyCode:         sql.NullString{String: t.IsoCurrencyCode, Valid: true},
			Date:                    sql.NullString{String: t.Date.Format("2006-01-02"), Valid: true},
			MerchantName:            sql.NullString{String: t.MerchantName, Valid: true},
			PaymentChannel:          t.PaymentChannel,
			PersonalFinanceCategory: t.PersonalFinanceCategory,
		}

		_, err = app.Config.Db.CreateTransaction(context.Background(), params)
		if err != nil {
			LogError(app.Config.Db, cmd, fmt.Errorf("error creating transaction record: %w", err), "Local database error")
			return err
		}

		fmt.Printf("\r%v", t.Id)
	}

	fmt.Println("\r > Transaction data fetched successfully.")

	return nil
}

// Gets all accounts for item
func (app *CLIApp) commandGetAccounts(cmd *cobra.Command, args []string) error {
	itemName := args[0]

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting credentials")
		return err
	}

	item, err := getItemFromServer(app, itemName)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting item")
		return err
	}

	accountsURL := app.Config.Client.BaseURL + "/api/items/" + item.ItemId + "/access/accounts"

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("POST", accountsURL, token, nil)
	})
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http request: %w", err), "Error contacting server")
		return err
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return err
	}

	var response models.Accounts

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("decoding err: %w", err), "Error contacting server")
		return err
	}

	if len(response.Accounts) == 0 {
		fmt.Printf("No accounts found for item: %s\n", itemName)
		return nil
	}

	for _, acc := range response.Accounts {
		avBalance := sql.NullFloat64{}
		if acc.AvailableBalance != "" {
			avBal, err := strconv.ParseFloat(acc.AvailableBalance, 64)
			if err != nil {
				LogError(app.Config.Db, cmd, fmt.Errorf("error converting string value: %w", err), "Data error")
				return err
			}
			avBalance.Float64 = avBal
			avBalance.Valid = true
		}

		curBalance := sql.NullFloat64{}
		if acc.CurrentBalance != "" {
			curBal, err := strconv.ParseFloat(acc.CurrentBalance, 64)
			if err != nil {
				LogError(app.Config.Db, cmd, fmt.Errorf("error converting string value: %w", err), "Data error")
				return err
			}
			curBalance.Float64 = curBal
			curBalance.Valid = true
		}

		params := database.UpsertAccountParams{
			ID:               acc.Id,
			CreatedAt:        time.Now().Format("2006-01-02"),
			UpdatedAt:        time.Now().Format("2006-01-02"),
			Name:             acc.Name,
			Type:             acc.Type,
			Subtype:          sql.NullString{String: acc.Subtype, Valid: true},
			Mask:             sql.NullString{String: acc.Mask, Valid: true},
			OfficialName:     sql.NullString{String: acc.OfficialName, Valid: true},
			AvailableBalance: avBalance,
			CurrentBalance:   curBalance,
			IsoCurrencyCode:  sql.NullString{String: acc.IsoCurrencyCode, Valid: true},
			InstitutionName:  sql.NullString{String: item.InstitutionName, Valid: true},
			UserID:           creds.User.ID.String(),
		}

		_, err := app.Config.Db.UpsertAccount(context.Background(), params)
		if err != nil {
			LogError(app.Config.Db, cmd, fmt.Errorf("error updating local account record: %w", err), "Local database error")
			return err
		}

	}

	fmt.Println(" > Account data fetched successfully.")

	return nil
}

// Rename an item
func (app *CLIApp) commandRenameItem(cmd *cobra.Command, args []string) error {
	itemCurrent := args[0]
	itemRename := args[1]

	item, err := getItemFromServer(app, itemCurrent)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting item")
		return err
	}

	renameURL := app.Config.Client.BaseURL + "/api/items/" + item.ItemId + "/name"

	request := models.UpdateItemName{
		Nickname: itemRename,
	}

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("PUT", renameURL, token, request)
	})
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http request: %w", err), "Error contacting server")
		return nil
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return nil
	}

	fmt.Printf("Item name successfully updated from %s to %s\n", itemCurrent, itemRename)
	return nil
}

// Deletes an item record for user
func (app *CLIApp) commandDeleteItem(cmd *cobra.Command, args []string) error {
	itemName := args[0]

	item, err := getItemFromServer(app, itemName)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting item")
		return err
	}

	pw, err := auth.ReadPassword("Please enter your password > ")
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting password")
		return nil
	}

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting credentials")
		return nil
	}

	user, err := app.Config.Db.GetUser(context.Background(), creds.User.Name)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error getting user record: %w", err), "Local database error")
		return nil
	}

	err = auth.ValidatePasswordHash(user.HashedPassword, pw)
	if err != nil {
		fmt.Println("Password entered is not correct")
		return nil
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println(" < Upon completion, all records for this item will be permanently deleted. > ")
	fmt.Println(" < If you wish to access this item's data in the future, you will have to link it once again.")
	fmt.Printf(" < Are you sure you want to delete item - %s? (y/n) > \n", itemName)
	for {
		fmt.Print(" > ")
		scanner.Scan()
		if scanner.Text() == "n" {
			fmt.Println("Item deletion aborted.")
			return nil
		} else if scanner.Text() == "y" {
			break
		} else {
			fmt.Println(" < Please enter either 'y' or 'n' > ")
		}
	}

	deleteURL := app.Config.Client.BaseURL + "/api/items/" + item.ItemId

	resp, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("DELETE", deleteURL, token, nil)
	})
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error making http request: %w", err), "Error contacting server")
		return nil
	}
	defer resp.Body.Close()

	err = checkResponseStatus(resp)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error contacting server")
		return nil
	}

	params := database.DeleteAccountsParams{
		InstitutionName: sql.NullString{String: item.InstitutionName, Valid: true},
		UserID:          user.ID,
	}
	err = app.Config.Db.DeleteAccounts(context.Background(), params)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error deleting account record: %w", err), "Local database error")
		return nil
	}

	fmt.Println("Item records deleted successfully")
	return nil
}

// Command resolves Plaid re-authentication issues through Link update flow. Calls sync command immediately afterwards
func (app *CLIApp) commandUpdate(cmd *cobra.Command, args []string) error {
	itemName := args[0]

	item, err := getItemFromServer(app, itemName)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting item")
		return err
	}

	err = linkUpdateModeFlow(app, item.ItemId)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error during financial institution re-authentication")
		return nil
	}

	err = app.commandSync(&cobra.Command{Use: "update"}, []string{itemName})
	if err != nil {
		return err
	}

	return nil
}

func (app *CLIApp) commandAddItem(cmd *cobra.Command) error {
	itemsURL := app.Config.Client.BaseURL + "/api/items"

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		LogError(app.Config.Db, cmd, err, "Error getting credentials")
		return nil
	}

	linked, err := userFirstTimePlaidLinkHelper(app, creds, itemsURL)
	if err != nil {
		LogError(app.Config.Db, cmd, fmt.Errorf("error linking: %w", err), "Error connecting financial institution")
		return nil
	}

	if !linked {
		fmt.Printf("Linking account to financial institution failed")
		return nil
	}

	return nil
}
