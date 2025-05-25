package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func (c *Config) ExitCurrentAccountSession() {
	
}

//Get all account names that exist for a user
func (c *Config) GetAccounts() ([]string, error) {

	accounts := make([]string, 0)
	accDir, err := c.getAccountsPath()
	if err != nil {
		return accounts, err
	}

	files, err := os.ReadDir(accDir)
	if err != nil {
		if os.IsNotExist(err) {
			return accounts, err 
		}
		return accounts, fmt.Errorf("error reading user accounts directory: %w", err)
	}

	if len(files) == 0 {
		return accounts, nil
	}

	for _, file := range files {
		accounts = append(accounts, file.Name())
	}

	return accounts, nil
}

//Create a new account under user
func (c *Config) CreateAccount() error {
	username := c.FileData.User.Name
	accountName := c.FileData.Account.Name
	fileData := c.FileData.Account

	configPath, err := getBaseConfigPath()
	if err != nil {
		return err
	}

	accDir := filepath.Join(configPath, "currentsession", username, "accounts", accountName)
	if err := os.MkdirAll(accDir, 0755); err != nil {
		return fmt.Errorf("error creating new account directory: %w", err)
	}

	filePath := filepath.Join(accDir, accountName+".json")
	
	jsonData, err := json.Marshal(fileData)
	if err != nil {
		return fmt.Errorf("erro marshalling json data: %w", err)
	}

	err = os.WriteFile(filePath, jsonData, 0600)
	if err != nil {
		return fmt.Errorf("error writing json file for account: %w", err)
	}

	return nil
}

//Get the path for user's accounts in the /users config directory
func (c *Config) getAccountsPath() (string, error) {
	configPath, err := getBaseConfigPath()
	if err != nil {
		return "", err
	}

	accountsDir := filepath.Join(configPath, "users",c.FileData.User.Name, "accounts")

	return accountsDir, nil
}


//Creates accounts subdirectory for a created user
func (c *Config) CreateAccountsConfigFilePath(username string) error {
	configPath, err := getBaseConfigPath()
	if err != nil {
		return err
	}

	accountsDir := filepath.Join(configPath, "users", username, "accounts")
	if err := os.MkdirAll(accountsDir, 0755); err != nil {
		return fmt.Errorf("error creating accounts directory for user: %w", err)
	}

	return nil
}