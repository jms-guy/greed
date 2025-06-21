package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jms-guy/greed/models"
)

//Store auth tokens in a credentials file 
func StoreTokens(data models.LoginResponse, configPath string) error {
	base, err := getBaseConfigPath(configPath)
	if err != nil {
		return err 
	}

	if err := os.MkdirAll(base, 0755); err != nil {
		return fmt.Errorf("error creating config dir: %w", err)
	}

	creds := filepath.Join(base, "credentials.json")

	f, err := os.OpenFile(creds, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
        return err
    }
    defer f.Close()

    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ") // optional, for pretty-print
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