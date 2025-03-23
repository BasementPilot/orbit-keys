// Package utils provides utility functions for working with API keys in the OrbitKeys system.
// It includes functions for generating, validating, and creating API keys with appropriate
// security measures.
package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/BasementPilot/orbit-keys/internal/models"
)

const (
	// DefaultKeyLength specifies the default length of generated API keys in bytes.
	// This provides a good balance between security and usability.
	DefaultKeyLength = 32
	
	// KeyPrefix is the string prefix added to all API keys for identification.
	// All valid API keys in the system will start with this prefix.
	KeyPrefix = "orbitkey_"
)

// GenerateAPIKey creates a new cryptographically secure API key with the specified length.
// The key is generated using secure random bytes and encoded using URL-safe base64.
// If length is <= 0, DefaultKeyLength will be used instead.
//
// The generated key will be prefixed with KeyPrefix and have any trailing '=' characters removed.
// Returns the generated key as a string and any error encountered during generation.
func GenerateAPIKey(length int) (string, error) {
	if length <= 0 {
		length = DefaultKeyLength
	}
	
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	
	key := base64.URLEncoding.EncodeToString(bytes)
	// Remove trailing = characters
	key = strings.TrimRight(key, "=")
	
	return KeyPrefix + key, nil
}

// ValidateAPIKey checks if a given string is a valid API key.
// It verifies the key starts with the correct prefix and has appropriate length.
//
// Returns true if the key is valid, false otherwise.
func ValidateAPIKey(key string) bool {
	// Check if key has the correct prefix
	if !strings.HasPrefix(key, KeyPrefix) {
		return false
	}
	
	// Check if the key is of appropriate length
	trimmedKey := strings.TrimPrefix(key, KeyPrefix)
	return len(trimmedKey) >= DefaultKeyLength
}

// IsRootAPIKey compares a provided key with the root API key to check for a match.
// It ensures both keys are non-empty and identical.
//
// Returns true if the key matches the root key, false otherwise.
func IsRootAPIKey(key, rootKey string) bool {
	return key == rootKey && key != ""
}

// CreateAPIKey generates a new API key and creates an APIKey model instance associated with a role.
// The key is cryptographically secure and follows the system's standard format.
//
// Parameters:
//   - roleID: The ID of the role to associate with this key
//   - description: A human-readable description of the key's purpose
//   - expiresIn: Optional duration after which the key will expire (nil for no expiration)
//
// Returns the created APIKey model and any error encountered during creation.
func CreateAPIKey(roleID uint, description string, expiresIn *time.Duration) (*models.APIKey, error) {
	// Generate new API key
	key, err := GenerateAPIKey(DefaultKeyLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}
	
	// Create new API key record
	apiKey := &models.APIKey{
		Key:         key,
		RoleID:      roleID,
		Description: description,
		CreatedAt:   time.Now(),
	}
	
	// Set expiration if provided
	if expiresIn != nil && *expiresIn > 0 {
		expiresAt := time.Now().Add(*expiresIn)
		apiKey.ExpiresAt = &expiresAt
	}
	
	return apiKey, nil
} 