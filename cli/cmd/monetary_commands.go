package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jms-guy/greed/cli/internal/auth"
	"github.com/jms-guy/greed/cli/internal/charts"
	"github.com/jms-guy/greed/cli/internal/database"
	//"github.com/jms-guy/greed/cli/internal/tables"
	"github.com/jms-guy/greed/models"
)

func (app *CLIApp) commandGetIncome(accountName string) error {

	creds, err := auth.GetCreds(app.Config.ConfigFP)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	user, err := app.Config.Db.GetUser(context.Background(), creds.User.Name)
	if err != nil {
		return fmt.Errorf("error getting local user record: %w", err)
	}

	params := database.GetAccountParams{
		Name: accountName,
		UserID: user.ID,
	}
	account, err := app.Config.Db.GetAccount(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error getting local account record: %w", err)
	}

	incURL := app.Config.Client.BaseURL + "/api/accounts/" + account.ID + "/transactions/monetary"

	res, err := DoWithAutoRefresh(app, func(token string) (*http.Response, error) {
		return app.Config.MakeBasicRequest("GET", incURL, token, nil) 
	})
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	checkResponseStatus(res)

	var response []models.MonetaryData
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	err = charts.MakeIncomeChart(response)
	if err != nil {
		return err
	}
	//tbl := tables.MakeTableForMonetaryAggregate(response, accountName)
	//tbl.Print()

	return nil
}