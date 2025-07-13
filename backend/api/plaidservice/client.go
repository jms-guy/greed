package plaidservice

import "github.com/plaid/plaid-go/v36/plaid"

// Creates a new APIClient for Plaid requests
func NewPlaidClient(clientID, secret string) *plaid.APIClient {
	config := plaid.NewConfiguration()
	config.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	config.AddDefaultHeader("PLAID-SECRET", secret)
	config.UseEnvironment(plaid.Production)
	client := plaid.NewAPIClient(config)
	return client
}