// Package utils provides utility functions for working with API keys in the OrbitKeys system.
// It includes functions for generating, validating, and creating API keys with appropriate
// security measures.
package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BasementPilot/orbit-keys/internal/models"
)

// Security-related errors
var (
	ErrKeyGeneration    = errors.New("failed to generate secure random bytes")
	ErrInvalidKeyLength = errors.New("invalid key length")
	ErrEmptyRootKey     = errors.New("root API key cannot be empty")
	ErrEmptyKey         = errors.New("API key cannot be empty")
)

const (
	// DefaultKeyLength specifies the default length of generated API keys in bytes.
	// 32 bytes (256 bits) provides strong security against brute force attacks.
	DefaultKeyLength = 32

	// MinKeyLength specifies the minimum acceptable length for API keys.
	// This ensures a baseline of security for all keys in the system.
	MinKeyLength = 16

	// KeyPrefix is the string prefix added to all API keys for identification.
	// All valid API keys in the system will start with this prefix.
	KeyPrefix = "orbitkey_"

	// MinTrimmedKeyLength is the minimum length a key should have after removing the prefix
	// This ensures keys have sufficient entropy
	MinTrimmedKeyLength = 22
)

// GenerateAPIKey creates a new cryptographically secure API key with the specified length.
// The key is generated using secure random bytes and encoded using URL-safe base64.
// If length is < MinKeyLength, DefaultKeyLength will be used instead.
//
// The generated key will be prefixed with KeyPrefix and have any trailing '=' characters removed.
// Returns the generated key as a string and any error encountered during generation.
func GenerateAPIKey(length int) (string, error) {
	if length < MinKeyLength {
		length = DefaultKeyLength
	}

	bytes := make([]byte, length)
	n, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrKeyGeneration, err)
	}

	// Verify we got the expected number of random bytes
	if n != length {
		return "", fmt.Errorf("%w: requested %d bytes but got %d", ErrInvalidKeyLength, length, n)
	}

	key := base64.URLEncoding.EncodeToString(bytes)
	// Remove trailing = characters
	key = strings.TrimRight(key, "=")

	return KeyPrefix + key, nil
}

// ValidateAPIKey checks if a given string is a valid API key.
// It verifies the key is not empty, starts with the correct prefix, and has appropriate length.
//
// Returns true if the key is valid, false otherwise.
func ValidateAPIKey(key string) bool {
	// Check if key is empty
	if key == "" {
		return false
	}

	// Check if key has the correct prefix
	if !strings.HasPrefix(key, KeyPrefix) {
		return false
	}

	// Check if the key is of appropriate length
	trimmedKey := strings.TrimPrefix(key, KeyPrefix)
	return len(trimmedKey) >= MinTrimmedKeyLength
}

// IsRootAPIKey compares a provided key with the root API key to check for a match.
// It ensures both keys are non-empty and uses a constant-time comparison to prevent timing attacks.
//
// Returns true if the key matches the root key, false otherwise.
// Will also return false if either key is empty.
func IsRootAPIKey(key, rootKey string) bool {
	// Both keys must be non-empty
	if key == "" || rootKey == "" {
		return false
	}

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(key), []byte(rootKey)) == 1
}

// CreateAPIKey generates a new API key and creates an APIKey model instance associated with a role.
// The key is cryptographically secure and follows the system's standard format.
//
// Parameters:
//   - roleID: The ID of the role to associate with this key
//   - description: A human-readable description of the key's purpose
//   - customData: Optional JSON string for storing custom metadata like user IDs
//   - expiresIn: Optional duration after which the key will expire (nil for no expiration)
//
// Returns the created APIKey model and any error encountered during creation.
func CreateAPIKey(roleID uint, description string, customData string, expiresIn *time.Duration) (*models.APIKey, error) {
	if roleID == 0 {
		return nil, errors.New("role ID cannot be zero")
	}

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
		CustomData:  customData,
		CreatedAt:   time.Now(),
	}

	// Set expiration if provided
	if expiresIn != nil && *expiresIn > 0 {
		// Cap maximum expiration time to reasonable limit (e.g., 10 years)
		maxDuration := 10 * 365 * 24 * time.Hour // 10 years
		duration := *expiresIn

		if duration > maxDuration {
			duration = maxDuration
		}

		expiresAt := time.Now().Add(duration)
		apiKey.ExpiresAt = &expiresAt
	}

	return apiKey, nil
}
