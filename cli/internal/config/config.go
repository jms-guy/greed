package config

import (
	"fmt"
	"net/url"
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
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		return nil, fmt.Errorf("SERVER_ADDRESS environment variable not set")
	}

	parsedAddress, err := url.Parse(serverAddress)
	if err != nil {
		return nil, fmt.Errorf("error parsing server address: %w", err)
	}
	if parsedAddress.Scheme != "http" && parsedAddress.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme in SERVER_ADDRESS: only 'http' and 'https' are allowed, got: %s", parsedAddress.Scheme)
	}

	client := NewClient(serverAddress)

	// Database path can either be absolute or just filename, if just a filename is provided, make sure
	// it's stored in home directory
	localDatabase := os.Getenv("LOCAL_DATABASE")
	if localDatabase == "" {
		return nil, fmt.Errorf("LOCAL_DATABASE environment variable not set")
	}
	var localDb string
	if filepath.IsAbs(localDatabase) {
		localDb = localDatabase
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error finding home directory: %w", err)
		}
		localDb = filepath.Join(homeDir, localDatabase)
	}

	queries, err := mySQL.OpenLocalDatabase(localDb)
	if err != nil {
		return nil, fmt.Errorf("error opening local database connection: %w", err)
	}

	cFP := os.Getenv("CONFIG_FILEPATH")
	if cFP == "" {
		cFP = ".config/greed" // Default
	}

	os := runtime.GOOS

	config := Config{
		Client:          client,
		Db:              queries,
		ConfigFP:        cFP,
		OperatingSystem: os,
	}

	return &config, nil
}
