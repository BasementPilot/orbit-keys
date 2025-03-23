package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	RootAPIKey string
	DBPath     string
	BaseURL    string
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	config := &Config{
		RootAPIKey: os.Getenv("ORBITKEYS_ROOT_API_KEY"),
		DBPath:     os.Getenv("ORBITKEYS_DB_PATH"),
		BaseURL:    os.Getenv("ORBITKEYS_BASE_URL"),
	}

	// Set default values if not provided
	if config.DBPath == "" {
		config.DBPath = "orbitkeys.db"
	} else {
		// Create directory if it doesn't exist
		dir := filepath.Dir(config.DBPath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Printf("Failed to create database directory: %v", err)
			}
		}
	}

	if config.BaseURL == "" {
		config.BaseURL = "/api"
	}

	return config
}

// SaveConfig saves the configuration to a .env file
func SaveConfig(config *Config) error {
	envContent := ""
	if config.RootAPIKey != "" {
		envContent += "ORBITKEYS_ROOT_API_KEY=" + config.RootAPIKey + "\n"
	}
	if config.DBPath != "" {
		envContent += "ORBITKEYS_DB_PATH=" + config.DBPath + "\n"
	}
	if config.BaseURL != "" {
		envContent += "ORBITKEYS_BASE_URL=" + config.BaseURL + "\n"
	}

	return os.WriteFile(".env", []byte(envContent), 0644)
}

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) bool {
	if config.RootAPIKey == "" {
		log.Println("Warning: ORBITKEYS_ROOT_API_KEY is not set. This is required for admin operations.")
		return false
	}
	return true
} 