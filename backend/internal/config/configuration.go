package config

import (
	"fmt"
	"os"
	"strconv"
)

// Configuration struct holding all .env variables for server
type Config struct {
	Port              string
	Environment       string
	DatabaseURL       string
	StaticAssetsPath  string
	JWTSecret         string
	JWTIssuer         string
	JWTAudience       string
	JWTExpiration     string // in seconds
	RefreshExpiration string // in seconds
	RateLimit         float64
	RateRefresh       float64 // per second
	SendGridAPIKey    string
	GreedEmail        string
	PlaidClientID     string
	PlaidSecret       string
	PlaidWebhookURL   string
	AESKey            string
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Printf("DEBUG: PORT env var = '%s'\n", port)
		port = "8080"
	}

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		return nil, fmt.Errorf("ENVIRONMENT variable not set")
	}

	dbURL := os.Getenv("DATABASE_URL")

	staticPath := os.Getenv("STATIC_ASSETS_PATH")
	if staticPath == "" {
		return nil, fmt.Errorf("STATIC_ASSETS_PATH environment variable not set")
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

	rateLimit := os.Getenv("RATE_LIMIT")
	if rateLimit == "" {
		return nil, fmt.Errorf("RATE_LIMIT environment variable not set")
	}

	limit, err := strconv.ParseFloat(rateLimit, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing RATE_LIMIT variable to flaot64: %w", err)
	}

	rateRefresh := os.Getenv("RATE_REFRESH")
	if rateRefresh == "" {
		return nil, fmt.Errorf("RATE_REFRESH environment variable not set")
	}

	refresh, err := strconv.ParseFloat(rateRefresh, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing RATE_REFRESH variable to float64: %w", err)
	}

	sendGridAPIKey := os.Getenv("SENDGRID_API_KEY")
	if sendGridAPIKey == "" {
		return nil, fmt.Errorf("SENDGRID_API_KEY environment variable not set")
	}

	greedEmail := os.Getenv("GREED_EMAIL")
	if greedEmail == "" {
		return nil, fmt.Errorf("GREED_EMAIL environment variable not set")
	}

	plaidClientID := os.Getenv("PLAID_CLIENT_ID")
	if plaidClientID == "" {
		return nil, fmt.Errorf("PLAID_CLIENT_ID environment variable not set")
	}

	plaidSecret := os.Getenv("PLAID_SECRET")
	if plaidSecret == "" {
		return nil, fmt.Errorf("PLAID_SECRET environment variable not set")
	}

	plaidWebhookURL := os.Getenv("PLAID_WEBHOOK_URL")
	if plaidWebhookURL == "" {
		return nil, fmt.Errorf("PLAID_WEBHOOK_URL environment variable not set")
	}

	aesKey := os.Getenv("AES_KEY")
	if aesKey == "" {
		return nil, fmt.Errorf("AES_KEY environment variable not set")
	}

	config := Config{
		Port:              port,
		Environment:       environment,
		DatabaseURL:       dbURL,
		StaticAssetsPath:  staticPath,
		JWTSecret:         jwtSecret,
		JWTIssuer:         jwtIssuer,
		JWTAudience:       jwtAudience,
		JWTExpiration:     jwtExpiration,
		RefreshExpiration: refreshExpiration,
		RateLimit:         limit,
		RateRefresh:       refresh,
		SendGridAPIKey:    sendGridAPIKey,
		GreedEmail:        greedEmail,
		PlaidClientID:     plaidClientID,
		PlaidSecret:       plaidSecret,
		PlaidWebhookURL:   plaidWebhookURL,
		AESKey:            aesKey,
	}

	return &config, nil
}
