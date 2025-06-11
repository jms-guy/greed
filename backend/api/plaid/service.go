package plaid

import (
	"context"
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

//Creates a new APIClient for Plaid requests
func NewPlaidClient(clientID, secret string) *plaid.APIClient {
	config := plaid.NewConfiguration()
	config.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	config.AddDefaultHeader("PLAID-SECRET", secret)
	config.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(config)
	return client
}


func GetLinkToken(client *plaid.APIClient, ctx context.Context, userID string) {
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: userID,
	}

	request := plaid.NewLinkTokenCreateRequest(
		"Greed-CLI",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_CA},
		user,
	)

	request.SetProducts([]plaid.Products{plaid.PRODUCTS_AUTH, plaid.PRODUCTS_BALANCE, plaid.PRODUCTS_TRANSACTIONS, plaid.PRODUCTS_IDENTITY})
	request.SetLinkCustomizationName("default")
	request.SetAccountFilters(plaid.LinkTokenAccountFilters{
		Depository: &plaid.DepositoryFilter{
			AccountSubtypes: []plaid.DepositoryAccountSubtype{plaid.DEPOSITORYACCOUNTSUBTYPE_CHECKING, plaid.DEPOSITORYACCOUNTSUBTYPE_SAVINGS},
		},
	})

	resp, _, err := client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()

	linkToken := resp.GetLinkToken()
}