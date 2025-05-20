package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
}

//Function moves user config file from .config directory to a current_session directory
//in order to manage currently logged in user
func (c *Config) SetCurrentSession(username string) error {
	//Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error finding home directory: %w", err)
	}

	//Get .env config file path
	configPath := os.Getenv("CONFIGFILEPATH")
	if configPath == "" {
		configPath = ".config/greed" //Default
	}

	//Make full current session directory path
	sessionDir := filepath.Join(homeDir, configPath, "currentsession")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
        return fmt.Errorf("error creating config directory: %w", err)
    }

	//Create file path
	newFilePath := filepath.Join(sessionDir, username+".json")

	//Get source file path
	oldFilePath, err := getConfigFilePath(username)
	if err != nil {
		return err
	}

	//Read user config file data
	data, err := os.ReadFile(oldFilePath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	//Write new config file in current session directory
	if err := os.WriteFile(newFilePath, data, 0600); err != nil {
		return fmt.Errorf("error writing to session file: %w", err)
	}

	//Remove .config directory file
	if err := os.Remove(oldFilePath); err != nil {
		return fmt.Errorf("error removing original config file: %w", err)
	}

	return nil
}

//Function moves a current session user config file back into the passive /users config directory
func (c *Config) EndCurrentSession() error {
	user := c.FileData.User.Name

	//Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error finding home directory: %w", err)
	}

	//Get .env config file path
	configPath := os.Getenv("CONFIGFILEPATH")
	if configPath == "" {
		configPath = ".config/greed" //Default
	}

	//Make full current session directory path
	sessionDir := filepath.Join(homeDir, configPath, "currentsession")
	configDir := filepath.Join(homeDir, configPath, "users")

	currFilePath := filepath.Join(sessionDir, user+".json")
	newFilePath := filepath.Join(configDir, user+".json")

	//Reads file
	data, err := os.ReadFile(currFilePath)
	if err != nil {
		return fmt.Errorf("error reading current session config file: %w", err)
	}

	//Writes new file
	if err := os.WriteFile(newFilePath, data, 0600); err != nil {
		return fmt.Errorf("error writing to users config directory: %w", err)
	}

	//Removes old file
	if err := os.Remove(currFilePath); err != nil {
		return fmt.Errorf("error removing session config file: %w", err)
	}

	return nil
}

//Loads the user config file located in the currentsession config directory into the Config struct
func (c *Config) LoadCurrentSession() error {
	//Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error finding home directory: %w", err)
	}

	//Get .env config path path
	configPath := os.Getenv("CONFIGFILEPATH")
	if configPath == "" {
		configPath = ".greed/config" //Default path
	}

	//Make full path
	sessionDir := filepath.Join(homeDir, configPath, "currentsession")

	//Read all files in current session directory
	files, err := os.ReadDir(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			//No active session
			return nil
		}
		return fmt.Errorf("error reading current session directory: %w", err)
	}

	//Check number of files in directory
	if len(files) == 0 {
		//No active session
		return nil
	}
	if len(files) > 1 {
		return	fmt.Errorf("multiple files found in current session directory")
	}

	//Get session file name
	sessionFile := filepath.Join(sessionDir, files[0].Name())

	//Read file
	file, err := os.Open(sessionFile)
	if err != nil {
		return fmt.Errorf("error opening session file: %w", err)
	}
	defer file.Close()

	//Read file
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading session file: %w", err)
	}

	//Unmarshal session data
	var data FileData
	if err := json.Unmarshal(fileBytes, &data); err != nil {
		return fmt.Errorf("error unmarshalling session data: %w", err)
	}

	c.FileData = data

	return nil
}

//Function creates a user config file in .config directory
func (c *Config) CreateUser(user models.User) error {
	//Create filedata struct to marshal into a config file
	fileData := FileData{User: user}

	//Find the config file for user (or create if new user)
	jsonFile, err := getConfigFilePath(user.Name)
	if err != nil {
		return fmt.Errorf("error getting config file path: %w", err)
	}

	//Marshal user into json data
	jsonData, err := json.Marshal(fileData)
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


//Get file path for config file, creates if does not already exist
func getConfigFilePath(username string) (string, error) {
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

	//Create config directory path
	configDir := filepath.Join(homeDir, configPath, "users")
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return "", fmt.Errorf("error creating config directory: %w", err)
    }

	//Create file path
	filePath := filepath.Join(configDir, username+".json")

	//Check if file exists, creates if not
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		_, err = os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0600)
		if err != nil {
			return "", fmt.Errorf("error creating config file: %w", err)
		}
	}

	return filePath, nil
}