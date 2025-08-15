package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jms-guy/greed/cli/internal/database"
	mySQL "github.com/jms-guy/greed/cli/sql"
)

// CLI config struct
type Config struct {
	Client          *Client           // Http client for handling server requests
	Db              *database.Queries // Local database queries
	ConfigFP        string            // Config file path
	OperatingSystem string            // Local operating system
}

// Initializes configuration struct
func LoadConfig() (*Config, error) {
	serverAddress := "https://greed-614554014047.us-central1.run.app"

	client := NewClient(serverAddress)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error finding home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "greed")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return nil, fmt.Errorf("error creating config directory: %w", err)
	}

	localDb := filepath.Join(configDir, "greed.db")

	queries, err := mySQL.OpenLocalDatabase(localDb)
	if err != nil {
		return nil, fmt.Errorf("error opening local database connection: %w", err)
	}

	os := runtime.GOOS

	config := Config{
		Client:          client,
		Db:              queries,
		ConfigFP:        ".config/greed",
		OperatingSystem: os,
	}

	return &config, nil
}
