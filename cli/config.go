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

//Function sets the user in the config file
func (c *Config) SetUser(user models.User) error {
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

//Reads config file data
func Read(userName string) (FileData, error) {
	//Get filepath for config file
	jsonFile, err := getConfigFilePath(userName)
	if err != nil {
		return FileData{}, err
	}

	//Open file
	file, err := os.Open(jsonFile)
	if err != nil {
		return FileData{}, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	//Read file
	fileBytes, _ := io.ReadAll(file)

	//Unmarshal data into struct
	var contents FileData
	if err := json.Unmarshal(fileBytes, &contents); err != nil {
		return FileData{}, fmt.Errorf("error retrieving json data: %w", err)
	}

	return contents, nil
}

//Get file path for config file, creates if does not already exist
func getConfigFilePath(userName string) (string, error) {
	//Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error finding home directory: %w", err)
	}

	//Get .env config file path
	configPath := os.Getenv("CONFIGFILEPATH")
	if configPath == "" {
		configPath = ".config/greed/users" //Default
	}

	//Create config directory path
	configDir := filepath.Join(homeDir, configPath)
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return "", fmt.Errorf("error creating config directory: %w", err)
    }

	//Create file path
	filePath := filepath.Join(configDir, userName+".json")

	//Check if file exists, creates if not
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		_, err = os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0600)
		if err != nil {
			return "", fmt.Errorf("error creating config file: %w", err)
		}
	}

	return filePath, nil
}