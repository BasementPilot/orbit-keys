package utils

import (
	"strings"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	// Test with default length
	key, err := GenerateAPIKey(0)
	if err != nil {
		t.Fatalf("GenerateAPIKey returned an error: %v", err)
	}
	
	// Check if key has the correct prefix
	if !strings.HasPrefix(key, KeyPrefix) {
		t.Errorf("Generated key %s does not have the prefix %s", key, KeyPrefix)
	}
	
	// Check if key has appropriate length
	trimmedKey := strings.TrimPrefix(key, KeyPrefix)
	if len(trimmedKey) < DefaultKeyLength {
		t.Errorf("Generated key %s is too short, expected at least %d characters", trimmedKey, DefaultKeyLength)
	}
	
	// Test with custom length
	customLength := 64
	key, err = GenerateAPIKey(customLength)
	if err != nil {
		t.Fatalf("GenerateAPIKey returned an error with custom length: %v", err)
	}
	
	// Check if key has the correct prefix
	if !strings.HasPrefix(key, KeyPrefix) {
		t.Errorf("Generated key %s does not have the prefix %s", key, KeyPrefix)
	}
}

func TestValidateAPIKey(t *testing.T) {
	// Test with valid key
	key, _ := GenerateAPIKey(DefaultKeyLength)
	if !ValidateAPIKey(key) {
		t.Errorf("ValidateAPIKey failed for a valid key: %s", key)
	}
	
	// Test with invalid prefix
	invalidKey := "invalid_prefix_" + strings.TrimPrefix(key, KeyPrefix)
	if ValidateAPIKey(invalidKey) {
		t.Errorf("ValidateAPIKey should have failed for an invalid key: %s", invalidKey)
	}
	
	// Test with too short key
	shortKey := KeyPrefix + "tooShort"
	if ValidateAPIKey(shortKey) {
		t.Errorf("ValidateAPIKey should have failed for a too short key: %s", shortKey)
	}
}

func TestIsRootAPIKey(t *testing.T) {
	rootKey := "orbitkey_root_test_key"
	
	// Test with matching keys
	if !IsRootAPIKey(rootKey, rootKey) {
		t.Errorf("IsRootAPIKey should return true for matching keys")
	}
	
	// Test with non-matching keys
	if IsRootAPIKey("different_key", rootKey) {
		t.Errorf("IsRootAPIKey should return false for non-matching keys")
	}
	
	// Test with empty key
	if IsRootAPIKey("", rootKey) {
		t.Errorf("IsRootAPIKey should return false for empty key")
	}
} 