package config 

import (
	"os"
	"fmt"
)

//Configuration struct holding all .env variables for server
type Config struct {
	ServerAddress 			string
	DatabaseURL				string
	JWTSecret				string
	JWTIssuer				string
	JWTAudience				string
	JWTExpiration			string		//in seconds
	RefreshExpiration		string		//in seconds
	SendGridAPIKey 			string
}

func LoadConfig() (*Config, error) {
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		return nil, fmt.Errorf("SERVER_ADDRESS environment variable not set")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable not set")
	}
	
	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		return nil, fmt.Errorf("JWT_ISSUER environment variable not set")
	}

	jwtAudience := os.Getenv("JWT_AUDIENCE")
	if jwtAudience == "" {
		return nil, fmt.Errorf("JWT_AUDIENCE environment variable not set")
	}

	jwtExpiration := os.Getenv("JWT_EXPIRATION")
	if jwtExpiration == "" {
		return nil, fmt.Errorf("JWT_EXPIRATION environment variable not set")
	}

	refreshExpiration := os.Getenv("REFRESH_EXPIRATION")
	if refreshExpiration == "" {
		return nil, fmt.Errorf("REFRESH_EXPIRATION environment variable not set")
	}

	sendGridAPIKey := os.Getenv("SENDGRID_API_KEY")
	if sendGridAPIKey == "" {
		return nil, fmt.Errorf("SENDGRID_API_KEY environment variable not set")
	}

	config := Config{
		ServerAddress: serverAddress,
		DatabaseURL: dbURL,
		JWTSecret: jwtSecret,
		JWTIssuer: jwtIssuer,
		JWTAudience: jwtAudience,
		JWTExpiration: jwtExpiration,
		RefreshExpiration: refreshExpiration,
		SendGridAPIKey: sendGridAPIKey,
	}

	return &config, nil
}