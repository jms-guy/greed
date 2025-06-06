package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/config"
)

//Function gets bearer authorization token from incoming request header
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
    if authHeader == "" {
        return "", fmt.Errorf("no Authorization header found")
    }
    
    // Remove any quotes that might be in the header
    authHeader = strings.Trim(authHeader, "\"")
    
    // Check if it starts with "Bearer "
    if !strings.HasPrefix(authHeader, "Bearer ") {
        return "", fmt.Errorf("authorization header format must be 'Bearer {token}'")
    }
    
    // Extract the token part
    token := strings.TrimPrefix(authHeader, "Bearer ")
    
    // Additional validation if needed
    if token == "" {
        return "", fmt.Errorf("token is empty")
    }
    
    return token, nil
}

//Function generates a JWT authorization token for a given userID
func MakeJWT(cfg *config.Config, userID uuid.UUID) (string, error) {

	//Get expiration time variable from .env
	expirationTime, err := strconv.Atoi(cfg.JWTExpiration)
	if err != nil {
		return "", fmt.Errorf("error getting JWT expiration time from .env: %w", err)
	}

	//Claims payload
	claims := jwt.RegisteredClaims{
		Issuer: 	cfg.JWTIssuer,
		IssuedAt: 	jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: 	jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(expirationTime) * time.Second)),
		NotBefore: 	jwt.NewNumericDate(time.Now().UTC()),
		Subject: 	userID.String(),
		ID:      	uuid.New().String(),
		Audience: 	jwt.ClaimStrings(strings.Split(cfg.JWTAudience, ",")),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("error creating JWT verification signature: %w", err)
	}

	return s, nil
}

//Validates a token input based off the secret string
func ValidateJWT(cfg *config.Config, tokenString string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	//Parse token string into claims token
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return []byte(cfg.JWTSecret), nil 
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("token is invalid or expired")
	}
	if !parsedToken.Valid {
		return uuid.UUID{}, fmt.Errorf("token is invalid")
	}

	if claims.Issuer != cfg.JWTIssuer {
		return uuid.UUID{}, fmt.Errorf("invalid issuer claim")
	}

	validAud := false
	for _, aud := range claims.Audience {
		if aud == "greed-cli-app" {
			validAud = true
			break
		}
	}
	if !validAud {
		return uuid.UUID{}, fmt.Errorf("invalid audience claim")
	}

	//Get userID subject string from claims 
	idString, err := parsedToken.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error getting user id from token: %w", err)
	}

	id, err := uuid.Parse(idString)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error parsing userID string: %w", err)
	}

	return id, nil
}