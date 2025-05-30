package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
)

//Interface to access database.Queries in auth package
type TokenStore interface {
	CreateToken(ctx context.Context, params database.CreateTokenParams) (database.RefreshToken, error)
}

const TokenLength = 32
const TokenExpirationTime = 2 * time.Hour

//Function creates a new token string, hashes the token, and creates a record of it in the database
func MakeRefreshToken(tokenStore TokenStore, delegationID uuid.UUID) (string, database.RefreshToken, error) {
	tokenString, err := generateRefreshToken()
	if err != nil {
		return "", database.RefreshToken{}, err 
	}

	hashedToken := HashRefreshToken(tokenString)

	token, err := storeRefreshToken(tokenStore, hashedToken, delegationID)
	if err != nil {
		return "", database.RefreshToken{}, err
	}

	return tokenString, token, nil
}

//Function creates a random 256-bit string key
func generateRefreshToken() (string, error) {
	key := make([]byte, TokenLength)
	
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("error creating token key: %w", err)
	}

	tokenString := hex.EncodeToString(key)
	return tokenString, nil
}

//Hash a refresh token using sha256
func HashRefreshToken(token string) string {
	hash := sha256.New()
	hash.Write([]byte(token))

	hashedToken := hex.EncodeToString(hash.Sum(nil))

	return hashedToken
}

//Creates refresh token record in database
func storeRefreshToken(tokenStore TokenStore, tokenHash string, delegationID uuid.UUID) (database.RefreshToken, error) {
	if tokenHash == "" {
        return database.RefreshToken{}, errors.New("token hash cannot be empty")
    }
    if delegationID == uuid.Nil {
        return database.RefreshToken{}, errors.New("delegation ID cannot be nil")
    }

	tokenParams := database.CreateTokenParams{
		ID: uuid.New(),
		HashedToken: tokenHash,
		DelegationID: delegationID,
		ExpiresAt: time.Now().Add(TokenExpirationTime),
	}

	token, err := tokenStore.CreateToken(context.Background(), tokenParams)
	if err != nil {
		return database.RefreshToken{}, fmt.Errorf("error creating refresh token in database: %w", err)
	}

	return token, nil
}