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
	// DefaultKeyLength is the default length of the API key in bytes
	DefaultKeyLength = 32
	
	// KeyPrefix is the prefix for API keys
	KeyPrefix = "orbitkey_"
)

// GenerateAPIKey generates a cryptographically secure API key
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

// ValidateAPIKey validates an API key format
func ValidateAPIKey(key string) bool {
	// Check if key has the correct prefix
	if !strings.HasPrefix(key, KeyPrefix) {
		return false
	}
	
	// Check if the key is of appropriate length
	trimmedKey := strings.TrimPrefix(key, KeyPrefix)
	return len(trimmedKey) >= DefaultKeyLength
}

// IsRootAPIKey checks if the provided key matches the root API key
func IsRootAPIKey(key, rootKey string) bool {
	return key == rootKey && key != ""
}

// CreateAPIKey creates a new API key associated with a role
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