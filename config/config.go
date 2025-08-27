package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Shopify struct {
		APIKey      string `json:"api_key"`
		APISecret   string `json:"api_secret"`
		Scopes      string `json:"scopes"`
		RedirectURI string `json:"redirect_uri"`
	} `json:"shopify"`
	Database struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		Name     string `json:"name"`
	} `json:"database"`
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
