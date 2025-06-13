package plaidservice

import (
	"context"

	"github.com/jms-guy/greed/models"
	"github.com/plaid/plaid-go/v36/plaid"
)

/*	Initial request
	1. Client - request server for link token
	2. Server - request link token from plaid - clientID/secret
	3. Server - send link token to client
	4. Client - opens link with plaid
	5. Client - OAuth/Plaid flow - gets public token from Plaid
	6. Server - receives public token from client
	7. Server - sends public token/clientID/secret to Plaid, gets back an access token
	8. Server - store access token in database (secure as any other financial information)

	Further requests
	1. Client - request server for data
	2. Server - send access token/clientID/secret to plaid
	3. Server - receive item from plaid
	4. Server - send item to client
	5. Client - display item contents
*/

// Creates a new APIClient for Plaid requests
func NewPlaidClient(clientID, secret string) *plaid.APIClient {
	config := plaid.NewConfiguration()
	config.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	config.AddDefaultHeader("PLAID-SECRET", secret)
	config.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(config)
	return client
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

	request.SetProducts([]plaid.Products{plaid.PRODUCTS_AUTH, plaid.PRODUCTS_BALANCE, plaid.PRODUCTS_TRANSACTIONS, plaid.PRODUCTS_IDENTITY})
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
func GetAccessToken(client *plaid.APIClient, ctx context.Context, sandboxPublicTokenResp plaid.SandboxPublicTokenCreateResponse) (models.AccessResponse, error) {

	exchangePublicTokenReq := plaid.NewItemPublicTokenExchangeRequest(sandboxPublicTokenResp.GetPublicToken())

	exchangePublicTokenResp, httpResp, err := client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*exchangePublicTokenReq,
	).Execute()
	if err != nil {
		return models.AccessResponse{}, err 
	}

	accessToken := exchangePublicTokenResp.GetAccessToken()
	itemID := exchangePublicTokenResp.ItemId
	reqID := httpResp.Header.Get("X-Request-Id")

	response := models.AccessResponse{
		AccessToken: accessToken,
		ItemID: itemID,
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

//Retrieves account data listed in a Plaid item
func GetAccounts(client *plaid.APIClient, ctx context.Context, accessToken string) ([]plaid.AccountBase, string, error){

	accountsGetRequest := plaid.NewAccountsGetRequest(accessToken)

	accountsGetResp, httpResp, err := client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		*accountsGetRequest,
	).Execute()
	if err != nil {
		return nil, "", err
	}

	accounts := accountsGetResp.GetAccounts()

	return accounts, httpResp.Header.Get("X-Request-Id"), nil
}