package plaidservice

import (
	"context"

	"github.com/jms-guy/greed/models"
	"github.com/plaid/plaid-go/v36/plaid"
)

type Service struct {
	Client *plaid.APIClient
}

// Plaid interface
type PlaidService interface {
	GetLinkToken(ctx context.Context, userID, webhookURL string) (string, error)
	GetLinkTokenForUpdateMode(ctx context.Context, userID, accessToken, webhookURL string) (string, error)
	GetAccessToken(ctx context.Context, publicToken string) (models.AccessResponse, error)
	InvalidateAccessToken(ctx context.Context, accessToken models.AccessResponse) (models.AccessResponse, error)
	GetAccounts(ctx context.Context, accessToken string) ([]plaid.AccountBase, string, error)
	GetItemInstitution(ctx context.Context, accessToken string) (string, error)
	GetBalances(ctx context.Context, accessToken string) (plaid.AccountsGetResponse, string, error)
	GetTransactions(ctx context.Context, accessToken, cursor string) (
		added, modified []plaid.Transaction, removed []plaid.RemovedTransaction, nextCursor, reqID string, err error)
	GetWebhookVerificationKey(ctx context.Context, keyID string) (plaid.JWKPublicKey, error)
	RemoveItem(ctx context.Context, accessToken string) error
}

// Creates a new APIClient for Plaid requests
func NewPlaidService(clientID, secret string) *Service {
	config := plaid.NewConfiguration()
	config.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	config.AddDefaultHeader("PLAID-SECRET", secret)
	config.UseEnvironment(plaid.Production)
	client := plaid.NewAPIClient(config)
	return &Service{Client: client}
}
