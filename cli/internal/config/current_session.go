package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//Function moves user config file from .config directory to a current_session directory
//in order to manage currently logged in user
func (c *Config) SetCurrentSession(username string) error {
	//Get base config path
	configPath, err := getBaseConfigPath()
	if err != nil {
		return err
	}

	//Make full current session directory path
	sessionDir := filepath.Join(configPath, "currentsession")

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

	//Get base config path
	configPath, err := getBaseConfigPath()
	if err != nil {
		return err
	}

	//Make full current session directory path
	sessionDir := filepath.Join(configPath, "currentsession")
	configDir := filepath.Join(configPath, "users")

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
	//Get base config path
	configPath, err := getBaseConfigPath()
	if err != nil {
		return err
	}

	//Make full path
	sessionDir := filepath.Join(configPath, "currentsession")

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