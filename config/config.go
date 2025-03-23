// Package config provides configuration management for the OrbitKeys system.
// It handles loading settings from environment variables or .env files, applying
// default values, and validating the configuration to ensure the system can operate properly.
package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config represents the application configuration settings.
// It contains all the parameters needed to run the OrbitKeys system.
type Config struct {
	// RootAPIKey is the master API key used for administrative operations.
	// If not provided, a new one will be generated during initialization.
	RootAPIKey string
	
	// DBPath is the path to the SQLite database file.
	// Defaults to "orbitkeys.db" in the current directory if not specified.
	DBPath string
	
	// BaseURL is the base URL prefix for all API endpoints.
	// Defaults to "/api" if not specified.
	BaseURL string
}

// LoadConfig reads configuration from environment variables and .env file.
// It tries to load variables from a .env file first, then falls back to environment variables.
// Default values are applied for any settings not explicitly provided.
//
// If the database directory specified in DBPath doesn't exist, it attempts to create it.
//
// Returns a populated Config struct with all settings.
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

// SaveConfig writes the configuration to a .env file.
// This is useful for persisting generated values like the root API key.
//
// Returns an error if writing to the file fails.
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

// ValidateConfig checks if the configuration has all required values set.
// Currently, it verifies that RootAPIKey is present, as this is essential for admin operations.
//
// Returns true if the configuration is valid, false otherwise with warning logs.
func ValidateConfig(config *Config) bool {
	if config.RootAPIKey == "" {
		log.Println("Warning: ORBITKEYS_ROOT_API_KEY is not set. This is required for admin operations.")
		return false
	}
	return true
} 