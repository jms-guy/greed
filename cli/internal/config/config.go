package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"github.com/jms-guy/greed/models"
)

//CLI config struct
type Config struct {
	FileData			FileData
	EnvData				EnvData
	Client				Client
}

//Struct containing .env file data
type EnvData struct {
	Address				string
}

//Struct containing .config file data
type FileData struct {
	User			models.User
	Account			models.Account
}

//Function creates a user config file in .config directory
func (c *Config) CreateUser(user models.User) error {
	//Find the config file for user (or create if new user)
	jsonFile, err := getConfigFilePath(user.Name)
	if err != nil {
		return fmt.Errorf("error getting config file path: %w", err)
	}

	//Marshal user into json data
	jsonData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("error marshalling json data: %w", err)
	}

	//Write the marshalled data into the config json file
	err = os.WriteFile(jsonFile, jsonData, 0600)
	if err != nil {
		return fmt.Errorf("error writing json file: %w", err)
	}

	return nil
}

//Function checks config files in .config/greed/users directory, and returns list of users client-side
func (c *Config) GetUsers() ([]string, error) {
	//Make empty users slice
	users := make([]string, 0)

	//Get base config path
	configPath, err := getBaseConfigPath()
	if err != nil {
		return users, err
	}

	configDir := filepath.Join(configPath, "users")
	
	//Read all files in config users directory
	files, err := os.ReadDir(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			//No users dir read 
			return users, fmt.Errorf("no user files found on machine")
		}
		return users, fmt.Errorf("error reading config users directory: %w", err)
	}

	//Check number of files in directory
	if len(files) == 0 {
		//No user files
		return users, nil
	}

	//Check each file, if {username.json} get username, add to users
	for _, file := range files {
		users = append(users, file.Name())
	}

	return users, nil
}

//Function will delete a user's config file on machine
func (c *Config) DeleteUser(username string) error {
	//Get base config path
	configPath, err := getBaseConfigPath()
	if err != nil {
		return err
	}

	filePath := filepath.Join(configPath, "currentsession", username)
	emptyConfigDir := filepath.Join(configPath, "users", username)

	//Remove user's .config file
	if err := os.RemoveAll(filePath); err != nil {
		return fmt.Errorf("error removing user's config directory: %w", err)
	}

	if err := os.Remove(emptyConfigDir); err != nil {
		return fmt.Errorf("error removing user's empty config dir: %w", err)
	}

	return nil
}


//Get file path for config file, creates if does not already exist
func getConfigFilePath(username string) (string, error) {
    configPath, err := getBaseConfigPath()
    if err != nil {
        return "", err
    }
    
    // Create config directory path
    configDir := filepath.Join(configPath, "users", username)
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return "", fmt.Errorf("error creating config directory: %w", err)
    }

    // Create file path
    filePath := filepath.Join(configDir, username+".json")

    // Check if file exists, creates if not
    if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
        _, err = os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0600)
        if err != nil {
            return "", fmt.Errorf("error creating config file: %w", err)
        }
    }

    return filePath, nil
}

//Gets the base "~/.config/greed" file path
func getBaseConfigPath() (string, error) {
	//Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error finding home directory: %w", err)
	}

	//Get .env config file path
	configPath := os.Getenv("CONFIGFILEPATH")
	if configPath == "" {
		configPath = ".config/greed" //Default
	}

	configDir := filepath.Join(homeDir, configPath)

	return configDir, nil
}
