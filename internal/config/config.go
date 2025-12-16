package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	APIBaseURL   string `mapstructure:"api_base_url"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	TokenPath    string `mapstructure:"token_path"`
}

func Load() (*Config, error) {
	setDefaults()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("api_base_url", "http://localhost:8085")
	viper.SetDefault("client_id", "7d8b1e7d-8c8d-4c7e-9f4a-2f0afc1a0f01")
	viper.SetDefault("client_secret", "cli-device-secret")

	tokenPath := defaultTokenPath()
	if tokenPath != "" {
		viper.SetDefault("token_path", tokenPath)
	}
}

func defaultTokenPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".config", "codestash", "token.json")
}
