package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server        ServerConfig   `json:"server"`
	Auth          AuthConfig     `json:"auth"`
	Database      DatabaseConfig `json:"database"`
	HomeAssistant HAConfig       `json:"home_assistant"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	User               string        `json:"user"`
	Password           string        `json:"password"`
	SecretKey          string        `json:"secret_key"`
	AccessTokenExpiry  time.Duration `json:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `json:"refresh_token_expiry"`
	NonceExpiry        time.Duration `json:"nonce_expiry"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Path string `json:"path"`
}

// HAConfig holds Home Assistant connection configuration
type HAConfig struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

var (
	instance *Config
	once     sync.Once
)

// LoadConfig loads configuration from the config file
func LoadConfig(configPath string) (*Config, error) {
	var err error
	once.Do(func() {
		instance = &Config{
			Server: ServerConfig{
				Host: "0.0.0.0",
				Port: 8080,
			},
			Auth: AuthConfig{
				User:               "admin",
				Password:           "admin",
				SecretKey:          "change-me-in-production-please",
				AccessTokenExpiry:  24 * time.Hour,      // 24 hours
				RefreshTokenExpiry: 30 * 24 * time.Hour, // 30 days
				NonceExpiry:        2 * time.Minute,     // 2 minutes
			},
			Database: DatabaseConfig{
				Path: "ha-mi.db",
			},
			HomeAssistant: HAConfig{
				URL:   "http://localhost:8123",
				Token: "",
			},
		}

		// If config file exists, load it
		if configPath != "" {
			absPath, err := filepath.Abs(configPath)
			if err != nil {
				err = fmt.Errorf("error getting absolute path: %w", err)
				return
			}

			file, err := os.Open(absPath)
			if err != nil {
				if os.IsNotExist(err) {
					// Create default config file if it doesn't exist
					saveErr := SaveConfig(configPath, instance)
					if saveErr != nil {
						err = fmt.Errorf("error creating default config: %w", saveErr)
					} else {
						err = nil
					}
					return
				}
				err = fmt.Errorf("error opening config file: %w", err)
				return
			}
			defer file.Close()

			decoder := json.NewDecoder(file)
			err = decoder.Decode(instance)
			if err != nil {
				err = fmt.Errorf("error decoding config file: %w", err)
				return
			}
		}
	})

	return instance, err
}

// SaveConfig saves the configuration to a file
func SaveConfig(configPath string, cfg *Config) error {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	file, err := os.Create(absPath)
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("error encoding config: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	if instance == nil {
		_, _ = LoadConfig("")
	}
	return instance
}
