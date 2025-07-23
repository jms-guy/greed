package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/plaid/plaid-go/v36/plaid"
)

//Struct that holds auth functions as methods
type Service struct {}


//Auth interface
type AuthService interface {
	EmailValidation(email string) bool
	ValidatePasswordHash(hash, password string) error
	HashPassword(password string) (string, error)
	GenerateCode() string
	GetBearerToken(headers http.Header) (string, error)
	MakeJWT(cfg *config.Config, userID uuid.UUID) (string, error)
	ValidateJWT(cfg *config.Config, tokenString string) (uuid.UUID, error)
	VerifyPlaidJWT(p PlaidKeyFetcher, ctx context.Context, tokenString string) error 
	MakeRefreshToken(tokenStore TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error)
	HashRefreshToken(token string) string
}

type PlaidKeyFetcher interface {
	GetWebhookVerificationKey(ctx context.Context, keyID string) (plaid.JWKPublicKey, error) 
}