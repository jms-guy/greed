package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

//Function creates a random 256-bit string key
func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("error creating token key: %w", err)
	}

	tokenString := hex.EncodeToString(key)
	return tokenString, nil
}