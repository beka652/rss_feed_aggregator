package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

var configFile = ".gatorconfig.json"

func Read() (*Config, error) {
	var config Config
	configFilePath, err  := getConfigFilePath()
	if err != nil {
		return  nil , err 
	}
	fileByte, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err 
	}
	err = json.Unmarshal(fileByte, &config)
	return &config, nil 
}

func (c *Config) SetUser() error {
	err := write(*c)
	if err != nil {
		return err 
	}
	return nil 
}

func write(cfg Config) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err 
	}
	configFile, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer configFile.Close() 
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "	")

	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}
	return nil 
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {	
		return "", err
	}
	filePath := filepath.Join(homeDir, configFile)
	return filePath, nil 
}