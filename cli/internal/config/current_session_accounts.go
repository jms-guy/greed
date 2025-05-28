package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"github.com/jms-guy/greed/models"
)

//Loads current session account data upon program start
//Looks into /.config/greed/currentsession/{user}/accounts, grabs account .json file from subdirectory
//and loads in into FileData
func (c *Config) LoadAccountData() error {
	configPath, err := getBaseConfigPath()
	if err != nil {
		return err 
	}

	if c.FileData.User.Name == "" {
		return nil
	}

	sessionDir := filepath.Join(configPath, "currentsession", c.FileData.User.Name, "accounts")

	dirs, err := os.ReadDir(sessionDir)
	if err != nil {
		return fmt.Errorf("error getting current session accounts dir: %w", err)
	}

	if len(dirs) == 0 {
		return nil
	}

	if len(dirs) != 1 {
		return fmt.Errorf("expecting a single account directory")
	}

	account := dirs[0].Name()

	accountSessionDir := filepath.Join(sessionDir, account)

	files, err := os.ReadDir(accountSessionDir)
	if err != nil {
		return fmt.Errorf("error reading current session directory: %w", err)
	}

	var accFile os.DirEntry
	accConfig := account + ".json"

	for _, file := range files {
		if file.Name() == accConfig {
			accFile = file
			break
		}
	}

	sessionFile := filepath.Join(accountSessionDir, accFile.Name())

	accountFile, err := os.Open(sessionFile)
	if err != nil {
		return fmt.Errorf("error opening session file for account: %w", err)
	}
	defer accountFile.Close()

	fileBytes, err := io.ReadAll(accountFile)
	if err != nil {
		return fmt.Errorf("error reading account session file: %w", err)
	}

	var data models.Account
	if err = json.Unmarshal(fileBytes, &data); err != nil {
		return fmt.Errorf("error unmarshalling account session data: %w", err)
	}

	c.FileData.Account = data

	return nil
}

//Updates an account json file with new data
func (c *Config) SaveAccountData() error {
	dataToSave := c.FileData.Account

	configDir, err := getBaseConfigPath()
	if err != nil {
		return fmt.Errorf("error getting base config path: %w", err)
	}

	accountConfig := c.FileData.Account.Name + ".json"
	sessionDir := filepath.Join(configDir, "currentsession", c.FileData.Account.Name)
	sessionFile := filepath.Join(sessionDir, accountConfig)

	data, err := json.Marshal(dataToSave)
	if err != nil {
		return fmt.Errorf("error marshalling save data: %w", err)
	}

	err = os.WriteFile(sessionFile, []byte(data), 0600)
	if err != nil {
		return fmt.Errorf("error writing save data to file: %w", err)
	}

	return nil
}

//Deletes a specific account directory, and all config files within
func (c *Config) DeleteAccountDir() error {
	configDir, err := getBaseConfigPath()
	if err != nil {
		return err 
	}

	accountPath := filepath.Join(configDir, "currentsession", c.FileData.User.Name, "accounts", c.FileData.Account.Name)

	if err = os.RemoveAll(accountPath); err != nil {
		return fmt.Errorf("error removing account data: %w", err)
	}

	return nil
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

//Moves a specific account directory to current session directory (logging into account)
func (c *Config) MoveAccountToCurrentSession(accountName string) error {
    configDir, err := getBaseConfigPath()
    if err != nil {
        return err
    }

    accountsDir, err := c.getAccountsPath()
    if err != nil {
        return err
    }

    accountPath := filepath.Join(accountsDir, accountName)
    currentSessionPath := filepath.Join(configDir, "currentsession", c.FileData.User.Name, "accounts", accountName)

    // Remove the destination if it already exists
    if _, err := os.Stat(currentSessionPath); err == nil {
        if err := os.RemoveAll(currentSessionPath); err != nil {
            return fmt.Errorf("error removing existing current session account dir: %w", err)
        }
    }

    if err = os.MkdirAll(filepath.Dir(currentSessionPath), 0755); err != nil {
        return fmt.Errorf("error making currentsession accounts dir: %w", err)
    }

    return os.Rename(accountPath, currentSessionPath)
}

//Moves a specific account directory to user's accounts directory (logging out of account)
func (c *Config) MoveAccountToUsersDir() error {
    configDir, err := getBaseConfigPath()
    if err != nil {
        return err
    }

    accountsDir, err := c.getAccountsPath()
    if err != nil {
        return err 
    }

    currentSessionPath := filepath.Join(configDir, "currentsession", c.FileData.User.Name, "accounts", c.FileData.Account.Name)
    accountPath := filepath.Join(accountsDir, c.FileData.Account.Name)

    // Remove the destination if it already exists
    if _, err := os.Stat(accountPath); err == nil {
        if err := os.RemoveAll(accountPath); err != nil {
            return fmt.Errorf("error removing existing user's account dir: %w", err)
        }
    }

    // Ensure parent directory exists
    if err = os.MkdirAll(filepath.Dir(accountPath), 0755); err != nil {
        return fmt.Errorf("error making user's accounts directory: %w", err)
    }

    return os.Rename(currentSessionPath, accountPath)
}