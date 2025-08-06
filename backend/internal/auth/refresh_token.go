package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
)

// Interface to access database.Queries in auth package
type TokenStore interface {
	CreateToken(ctx context.Context, params database.CreateTokenParams) (database.RefreshToken, error)
}

// Function creates a new token string, hashes the token, and creates a record of it in the database
func (s *Service) MakeRefreshToken(tokenStore TokenStore, userID uuid.UUID, delegation database.Delegation) (string, error) {
	tokenString, err := generateRefreshToken()
	if err != nil {
		return "", err
	}

	hashedToken := s.HashRefreshToken(tokenString)

	_, err = storeRefreshToken(tokenStore, hashedToken, userID, delegation)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Function creates a random 256-bit string key
func generateRefreshToken() (string, error) {
	const TokenLength = 32

	key := make([]byte, TokenLength)

	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("error creating token key: %w", err)
	}

	tokenString := hex.EncodeToString(key)
	return tokenString, nil
}

// Hash a refresh token using sha256
func (s *Service) HashRefreshToken(token string) string {
	hash := sha256.New()
	hash.Write([]byte(token))

	hashedToken := hex.EncodeToString(hash.Sum(nil))

	return hashedToken
}

// Creates refresh token record in database
func storeRefreshToken(tokenStore TokenStore, tokenHash string, userID uuid.UUID, delegation database.Delegation) (database.RefreshToken, error) {
	if tokenHash == "" {
		return database.RefreshToken{}, errors.New("token hash cannot be empty")
	}
	if delegation.ID == uuid.Nil {
		return database.RefreshToken{}, errors.New("delegation ID cannot be nil")
	}

	tokenParams := database.CreateTokenParams{
		ID:           uuid.New(),
		HashedToken:  tokenHash,
		UserID:       userID,
		DelegationID: delegation.ID,
		ExpiresAt:    delegation.ExpiresAt,
	}

	token, err := tokenStore.CreateToken(context.Background(), tokenParams)
	if err != nil {
		return database.RefreshToken{}, fmt.Errorf("error creating refresh token in database: %w", err)
	}

	return token, nil
}
