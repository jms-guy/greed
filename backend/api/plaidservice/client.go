package plaidservice

import "github.com/plaid/plaid-go/v36/plaid"

type PlaidService struct {
	Client 		*plaid.APIClient
}

// Creates a new APIClient for Plaid requests
func NewPlaidService(clientID, secret string) *PlaidService {
	config := plaid.NewConfiguration()
	config.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	config.AddDefaultHeader("PLAID-SECRET", secret)
	config.UseEnvironment(plaid.Production)
	client := plaid.NewAPIClient(config)
	return &PlaidService{Client: client}
}