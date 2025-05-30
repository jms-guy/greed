package auth

import (
	"os"
	"strings"
)

//JWT configuration from .env variables
var (
	JWTIssuer		= getEnvOrDefault("JWT_ISSUER", "greed")
	JWTAudience		= strings.Split(getEnvOrDefault("JWT_AUDIENCE", "greed-cli-app"), ",")
	JWTExpiration	= os.Getenv("JWT_EXPIRATION_SECONDS")
	JWTSecret		= os.Getenv("JWT_SECRET")
)

func getEnvOrDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val 
}
