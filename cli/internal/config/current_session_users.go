package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"github.com/jms-guy/greed/models"
)

//Saves current FileData struct to the current session user config file. 
func (c *Config) SaveUserFileData() error {
	dataToSave := c.FileData.User

	configPath, err := getBaseConfigPath()
	if err != nil {
		return fmt.Errorf("error getting base config path: %w", err)
	}

	userConfig := c.FileData.User.Name + ".json"
	sessionDir := filepath.Join(configPath, "currentsession", c.FileData.User.Name)
	sessionFile := filepath.Join(sessionDir, userConfig)

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

//Function moves user config file from .config directory to a current_session directory
//in order to manage currently logged in user
func (c *Config) SetCurrentUserSession(username string) error {
	//Get base config path
	configPath, err := getBaseConfigPath()
	if err != nil {
		return err
	}

	//Make full current session directory path
	sessionDir := filepath.Join(configPath, "currentsession", username)

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

	currAccountsDir := filepath.Join(sessionDir, "accounts")
	if err = os.Mkdir(currAccountsDir, 0755); err != nil {
		return fmt.Errorf("error creating current accounts directory: %w", err)
	}

	//Remove .config directory file
	if err := os.Remove(oldFilePath); err != nil {
		return fmt.Errorf("error removing original config file: %w", err)
	}

	return nil
}

//Function moves a current session user config file back into the passive /users config directory
func (c *Config) EndCurrentUserSession() error {
	user := c.FileData.User.Name

	//Get base config path
	configPath, err := getBaseConfigPath()
	if err != nil {
		return err
	}

	//Make full current session directory path
	sessionDir := filepath.Join(configPath, "currentsession", user)
	configDir := filepath.Join(configPath, "users", user)

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

	///Check if accounts dir in current session is empty, remove directory. Else move account back as well, then delete

	//Removes old file
	if err := os.Remove(currFilePath); err != nil {
		return fmt.Errorf("error removing session config file: %w", err)
	}

	if err := os.Remove(sessionDir); err != nil {
		return fmt.Errorf("error removing user's current session directory: %w", err)
	}

	return nil
}

//Loads the user config file located in the currentsession config directory into the Config struct
func (c *Config) LoadCurrentUserSession() error {
	//Get base config path
	configPath, err := getBaseConfigPath()
	if err != nil {
		return err
	}

	//Make full path
	sessionDir := filepath.Join(configPath, "currentsession")

	dirs, err := os.ReadDir(sessionDir)
	if err != nil {
		return fmt.Errorf("error getting current session user dir: %w", err)
	}

	if len(dirs) == 0 {
		return nil
	}

	if len(dirs) != 1 {
		return fmt.Errorf("expecting a single current session user directory")
	}

	user := dirs[0].Name()

	userSessionDir := filepath.Join(sessionDir, user)

	//Read all files in current session directory
	files, err := os.ReadDir(userSessionDir)
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

	var userFile os.DirEntry

	for _, file := range files {
		if file.Name() == user+".json"	{
			userFile = file
			break
		}
	}

	//Get session file name
	sessionFile := filepath.Join(userSessionDir, userFile.Name())

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
	var data models.User
	if err := json.Unmarshal(fileBytes, &data); err != nil {
		return fmt.Errorf("error unmarshalling session data: %w", err)
	}

	c.FileData.User = data

	return nil
}