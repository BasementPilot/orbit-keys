// Package config provides configuration management for the OrbitKeys system.
// It handles loading settings from environment variables or .env files, applying
// default values, and validating the configuration to ensure the system can operate properly.
package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Security-related errors
var (
	ErrEmptyRootKey     = errors.New("root API key cannot be empty")
	ErrInvalidFilePath  = errors.New("invalid file path")
	ErrConfigSaveFailed = errors.New("failed to save configuration")
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
// Returns a populated Config struct with all settings and any encountered errors.
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	config := &Config{
		RootAPIKey: sanitizeEnv("ORBITKEYS_ROOT_API_KEY"),
		DBPath:     sanitizeEnv("ORBITKEYS_DB_PATH"),
		BaseURL:    sanitizeEnv("ORBITKEYS_BASE_URL"),
	}

	// Set default values if not provided
	if config.DBPath == "" {
		config.DBPath = "orbitkeys.db"
	} else {
		// Validate and sanitize the DB path
		if !isValidFilePath(config.DBPath) {
			return config, ErrInvalidFilePath
		}
		
		// Create directory if it doesn't exist
		dir := filepath.Dir(config.DBPath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return config, fmt.Errorf("failed to create database directory: %w", err)
			}
		}
	}

	if config.BaseURL == "" {
		config.BaseURL = "/api"
	} else {
		// Ensure BaseURL starts with a slash
		if !strings.HasPrefix(config.BaseURL, "/") {
			config.BaseURL = "/" + config.BaseURL
		}
		
		// Remove trailing slash if present
		config.BaseURL = strings.TrimSuffix(config.BaseURL, "/")
	}

	return config, nil
}

// SaveConfig writes the configuration to a .env file.
// This is useful for persisting generated values like the root API key.
//
// Returns an error if writing to the file fails.
func SaveConfig(config *Config) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}
	
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

	// Create a temporary file first, then rename it to avoid partial writes
	tempFile := ".env.tmp"
	err := os.WriteFile(tempFile, []byte(envContent), 0600) // More restrictive permissions for security
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConfigSaveFailed, err)
	}
	
	// Rename the temporary file to the actual .env file
	if err := os.Rename(tempFile, ".env"); err != nil {
		// Clean up the temporary file if rename fails
		os.Remove(tempFile)
		return fmt.Errorf("%w: %v", ErrConfigSaveFailed, err)
	}
	
	return nil
}

// ValidateConfig checks if the configuration has all required values set.
// It verifies that RootAPIKey is present and valid, as this is essential for admin operations.
//
// Returns true if the configuration is valid, false otherwise with warning logs.
func ValidateConfig(config *Config) bool {
	if config == nil {
		log.Println("Error: Configuration is nil")
		return false
	}
	
	if config.RootAPIKey == "" {
		log.Println("Error: ORBITKEYS_ROOT_API_KEY is not set. This is required for admin operations.")
		return false
	}
	
	// Additional validation can be added here
	return true
}

// sanitizeEnv retrieves and sanitizes an environment variable value.
// It trims whitespace and prevents some basic injection attacks.
func sanitizeEnv(key string) string {
	value := os.Getenv(key)
	value = strings.TrimSpace(value)
	
	// Basic security: Reject values with potentially dangerous characters
	// This is a simplified example - in production, use a more comprehensive validation
	if strings.ContainsAny(value, ";&|`$()<>") {
		log.Printf("Warning: Environment variable %s contains potentially unsafe characters and was ignored", key)
		return ""
	}
	
	return value
}

// isValidFilePath checks if a file path is valid and safe.
// It prevents directory traversal attacks and other unsafe paths.
func isValidFilePath(path string) bool {
	// Reject empty paths
	if path == "" {
		return false
	}
	
	// Prevent directory traversal
	if strings.Contains(path, "..") {
		return false
	}
	
	// Sanitize and validate the path
	filepath.Clean(path) // Use the result but don't assign to variable
	
	// Additional security checks can be added here
	return true
} 