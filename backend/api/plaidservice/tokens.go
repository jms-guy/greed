package plaidservice

import (
	"context"
	"fmt"

	"github.com/jms-guy/greed/models"
	"github.com/plaid/plaid-go/v36/plaid"
)

//Sandbox access token generation
func GetSandboxToken(client *plaid.APIClient, ctx context.Context) (plaid.ItemPublicTokenExchangeResponse, error) {
	sandboxPublicTokenResp, _, err := client.PlaidApi.SandboxPublicTokenCreate(ctx).SandboxPublicTokenCreateRequest(
		*plaid.NewSandboxPublicTokenCreateRequest(
			"ins_109508",
			[]plaid.Products{plaid.PRODUCTS_TRANSACTIONS},
		),
	).Execute()
	if err != nil {
		if apiErr, ok := err.(plaid.GenericOpenAPIError); ok {
			fmt.Println("Plaid error body:", string(apiErr.Body()))
		}
		return plaid.ItemPublicTokenExchangeResponse{}, err
	}
	exchangePublicTokenResp, _, err := client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*plaid.NewItemPublicTokenExchangeRequest(sandboxPublicTokenResp.GetPublicToken()),
	  ).Execute()
	if err != nil {
		if apiErr, ok := err.(plaid.GenericOpenAPIError); ok {
			fmt.Println("Plaid error body:", string(apiErr.Body()))
		}
		return plaid.ItemPublicTokenExchangeResponse{}, err 
	}

	return exchangePublicTokenResp, nil
}

//Requests Plaid API for a Link token for client use
func GetLinkToken(client *plaid.APIClient, ctx context.Context, userID string) (string, error) {
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: userID,
	}

	request := plaid.NewLinkTokenCreateRequest(
		"Greed-CLI",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_CA},
		user,
	)

	transactions := plaid.LinkTokenTransactions{
		DaysRequested: plaid.PtrInt32(730),
	  }

	request.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})
	request.SetLinkCustomizationName("default")
	request.SetTransactions(transactions)
	request.SetAccountFilters(plaid.LinkTokenAccountFilters{
		Depository: &plaid.DepositoryFilter{
			AccountSubtypes: []plaid.DepositoryAccountSubtype{plaid.DEPOSITORYACCOUNTSUBTYPE_CHECKING, plaid.DEPOSITORYACCOUNTSUBTYPE_SAVINGS},
		},
		Credit: &plaid.CreditFilter{
			AccountSubtypes: []plaid.CreditAccountSubtype{plaid.CREDITACCOUNTSUBTYPE_CREDIT_CARD},
		},
	})

	resp, _, err := client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		return "", err
	}

	linkToken := resp.GetLinkToken()

	return linkToken, nil
}

//Exchanges a public token received from client for a permanent access token for item from Plaid API
func GetAccessToken(client *plaid.APIClient, ctx context.Context, publicToken string) (models.AccessResponse, error) {

	exchangePublicTokenReq := plaid.NewItemPublicTokenExchangeRequest(publicToken)

	exchangePublicTokenResp, httpResp, err := client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*exchangePublicTokenReq,
	).Execute()
	if err != nil {
		return models.AccessResponse{}, err 
	}

	accessToken := exchangePublicTokenResp.GetAccessToken()
	itemID := exchangePublicTokenResp.ItemId
	instName, err := GetItemInstitution(client, ctx, accessToken)
	if err != nil {
		return models.AccessResponse{}, err
	}

	reqID := httpResp.Header.Get("X-Request-Id")

	response := models.AccessResponse{
		AccessToken: accessToken,
		ItemID: itemID,
		InstitutionName: instName,
		RequestID: reqID,
	}

	return response, nil
}

//Invalidates an item's access token, requesting a new one from Plaid API
func InvalidateAccessToken(client *plaid.APIClient,ctx context.Context, accessToken models.AccessResponse) (models.AccessResponse, error) {

	request := plaid.NewItemAccessTokenInvalidateRequest(accessToken.AccessToken)

	response, httpResp, err := client.PlaidApi.ItemAccessTokenInvalidate(ctx).ItemAccessTokenInvalidateRequest(
		*request,
	).Execute()
	if err != nil {
		return models.AccessResponse{}, err
	}

	newAccessToken := response.GetNewAccessToken()
	itemID := accessToken.ItemID
	reqID := httpResp.Header.Get("X-Request-Id")

	newToken := models.AccessResponse{
		AccessToken: newAccessToken,
		ItemID: itemID,
		RequestID: reqID,
	}

	return newToken, nil
}