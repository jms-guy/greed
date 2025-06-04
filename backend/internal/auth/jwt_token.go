package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"os"
	"strings"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
func MakeJWT(userID uuid.UUID) (string, error) {
	//JWT configuration from .env variables
	JWTIssuer		:= getEnvOrDefault("JWT_ISSUER", "greed")
	JWTAudience		:= strings.Split(getEnvOrDefault("JWT_AUDIENCE", "greed-cli-app"), ",")
	JWTExpiration	:= os.Getenv("JWT_EXPIRATION_SECONDS")
	JWTSecret		:= os.Getenv("JWT_SECRET")

	//Get expiration time variable from .env
	expirationTime, err := strconv.Atoi(JWTExpiration)
	if err != nil {
		return "", fmt.Errorf("error getting JWT expiration time from .env: %w", err)
	}

	//Claims payload
	claims := jwt.RegisteredClaims{
		Issuer: 	JWTIssuer,
		IssuedAt: 	jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: 	jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(expirationTime) * time.Second)),
		NotBefore: 	jwt.NewNumericDate(time.Now().UTC()),
		Subject: 	userID.String(),
		ID:      	uuid.New().String(),
		Audience: 	jwt.ClaimStrings(JWTAudience),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", fmt.Errorf("error creating JWT verification signature: %w", err)
	}

	return s, nil
}

//Validates a token input based off the secret string
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	JWTIssuer		:= getEnvOrDefault("JWT_ISSUER", "greed")
	claims := &jwt.RegisteredClaims{}
	//Parse token string into claims token
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil 
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("token is invalid or expired")
	}
	if !parsedToken.Valid {
		return uuid.UUID{}, fmt.Errorf("token is invalid")
	}

	if claims.Issuer != JWTIssuer {
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