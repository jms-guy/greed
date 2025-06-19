package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

//CLI config struct
type Config struct {
	Client				*Client			//Http client for handling server requests
	LocalDatabase 		string 			//Filepath for accessing local database
	OperatingSystem 	string 			//Local operating system
}

//Initializes configuration struct 
func LoadConfig() (*Config, error) {
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		return nil, fmt.Errorf("SERVER_ADDRESS environment variable not set")
	}

	client := NewClient(serverAddress)

	//Database path can either be absolute or just filename, if just a filename is provided, make sure 
	//it's stored in home directory
	localDatabase := os.Getenv("LOCAL_DATABASE")
	if localDatabase == "" {
		return nil, fmt.Errorf("LOCAL_DATABASE environment variable not set")
	}
	var localDb string
	if filepath.IsAbs(localDatabase) {
		localDb = localDatabase
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error finding home directory: %w", err)
		}
		localDb = filepath.Join(homeDir, localDatabase)
	}

	os := runtime.GOOS


	config := Config{
		Client: 			client,
		LocalDatabase: 		localDb,
		OperatingSystem: 	os,
	}

	return &config, nil
}


