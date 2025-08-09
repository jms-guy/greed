package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jms-guy/greed/models"
)

// Remove credentials from local machine
func RemoveCreds(configPath string) error {
	base, err := getBaseConfigPath(configPath)
	if err != nil {
		return fmt.Errorf("error getting config directory: %w", err)
	}

	credFile := filepath.Join(base, "credentials.json")

	err = os.Remove(credFile)
	if err != nil {
		return fmt.Errorf("error removing credentials file: %w", err)
	}

	return nil
}

// Get credentials from JSON file
func GetCreds(configPath string) (models.Credentials, error) {
	var creds models.Credentials

	base, err := getBaseConfigPath(configPath)
	if err != nil {
		return creds, fmt.Errorf("error getting config directory: %w", err)
	}

	credFile := filepath.Join(base, "credentials.json")

	// #nosec G304 - configPath is trusted input
	file, err := os.Open(credFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return creds, fmt.Errorf("file does not exist")
		} else {
			return creds, fmt.Errorf("error opening cred file: %w", err)
		}
	}
	defer file.Close()

	byteVal, err := io.ReadAll(file)
	if err != nil {
		return creds, fmt.Errorf("error reading credentials.json: %w", err)
	}

	err = json.Unmarshal(byteVal, &creds)
	if err != nil {
		return creds, fmt.Errorf("error unmarshalling data: %w", err)
	}

	return creds, nil
}

// Store auth tokens in a credentials file
func StoreTokens(data models.Credentials, configPath string) error {
	base, err := getBaseConfigPath(configPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(base, 0o700); err != nil {
		return fmt.Errorf("error creating config dir: %w", err)
	}

	creds := filepath.Join(base, "credentials.json")

	// #nosec G304 - configPath is trusted input
	f, err := os.OpenFile(filepath.Clean(creds), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func getBaseConfigPath(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error finding home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, path) //.config/greed

	return configDir, nil
}
