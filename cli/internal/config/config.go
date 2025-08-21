package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jms-guy/greed/cli/internal/database"
	mySQL "github.com/jms-guy/greed/cli/sql"
	"github.com/jms-guy/greed/models"
)

// CLI config struct
type Config struct {
	Client          *Client           // Http client for handling server requests
	Db              *database.Queries // Local database queries
	ConfigFP        string            // Config file path
	OperatingSystem string            // Local operating system
	SettingsFP      string            // Settings filepath
	Settings        Settings          // Holds settings loaded from config file
}

// Settings loaded from config file
type Settings struct {
	DefaultItem    models.ItemName  `json:"default_item"`
	DefaultAccount database.Account `json:"default_account"`
}

// Initializes configuration struct
func LoadConfig() (*Config, error) {
	serverAddress := "https://greed-614554014047.us-central1.run.app"

	client := NewClient(serverAddress)

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error finding home directory: %w", err)
	}

	// Config directory
	configDir := filepath.Join(homeDir, ".config", "greed")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		return nil, fmt.Errorf("error creating config directory: %w", err)
	}

	// Settings
	settingsFile := filepath.Join(configDir, "settings.json")
	file, err := os.OpenFile(settingsFile, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, fmt.Errorf("error opening settings file: %w", err)
	}
	defer file.Close()

	settings, err := loadSettingsFromFile(settingsFile)
	if err != nil {
		return nil, fmt.Errorf("error loading settings from file")
	}

	// Local database file
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
		SettingsFP:      settingsFile,
		Settings:        settings,
	}

	return &config, nil
}

// Function to load settings from file
func loadSettingsFromFile(filePath string) (Settings, error) {
	var settings Settings

	file, err := os.Open(filePath)
	if err != nil {
		return settings, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&settings)
	if err != nil {
		if err == io.EOF {
			return settings, nil
		}
		return settings, fmt.Errorf("error decoding settings file: %w", err)
	}

	return settings, nil
}

// Function to save settings to file
func SaveSettingsToFile(filePath string, settings Settings) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(settings); err != nil {
		return err
	}

	return nil
}
