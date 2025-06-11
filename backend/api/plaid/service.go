package plaid

import "github.com/plaid/plaid-go/v36/plaid"

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

func NewPlaidClient(clientID, secret string) *plaid.APIClient {
	config := plaid.NewConfiguration()
	config.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	config.AddDefaultHeader("PLAID-SECRET", secret)
	config.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(config)
	return client
}