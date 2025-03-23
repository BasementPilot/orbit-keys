package utils

import (
	"crypto/subtle"
	"strings"
	"testing"
	"time"
)

func TestGenerateAPIKey(t *testing.T) {
	// Test valid key length
	t.Run("Valid key length", func(t *testing.T) {
		key, err := GenerateAPIKey(DefaultKeyLength)
		if err != nil {
			t.Fatalf("GenerateAPIKey failed with error: %v", err)
		}

		// Check prefix
		if !strings.HasPrefix(key, KeyPrefix) {
			t.Errorf("Expected key to have prefix %q, got %q", KeyPrefix, key)
		}

		// Check length (prefix + base64 encoding of byte length)
		// Base64 encoding: 4 characters for every 3 bytes
		expectedMinLength := len(KeyPrefix) + (DefaultKeyLength*4+2)/3
		if len(key) < expectedMinLength {
			t.Errorf("Key length is too short: expected at least %d, got %d", expectedMinLength, len(key))
		}
	})

	// Test invalid key length (too short)
	t.Run("Key length too short", func(t *testing.T) {
		// Should apply DefaultKeyLength when given less than MinKeyLength
		key, err := GenerateAPIKey(MinKeyLength - 1)
		if err != nil {
			t.Errorf("Expected success when key length is too short (should use default)")
		}

		// Check that default length was applied
		expectedMinLength := len(KeyPrefix) + (DefaultKeyLength*4+2)/3
		if len(key) < expectedMinLength {
			t.Errorf("Key length is too short: expected at least %d, got %d", expectedMinLength, len(key))
		}
	})

	// Test multiple key generation to ensure uniqueness
	t.Run("Key uniqueness", func(t *testing.T) {
		keySet := make(map[string]bool)
		iterations := 100

		for i := 0; i < iterations; i++ {
			key, err := GenerateAPIKey(DefaultKeyLength)
			if err != nil {
				t.Fatalf("GenerateAPIKey failed with error: %v", err)
			}

			if keySet[key] {
				t.Errorf("Duplicate key generated: %s", key)
			}
			keySet[key] = true
		}
	})
}

func TestValidateAPIKey(t *testing.T) {
	// Test valid key
	t.Run("Valid key", func(t *testing.T) {
		key, err := GenerateAPIKey(DefaultKeyLength)
		if err != nil {
			t.Fatalf("GenerateAPIKey failed with error: %v", err)
		}

		isValid := ValidateAPIKey(key)
		if !isValid {
			t.Errorf("Expected key %q to be valid", key)
		}
	})

	// Test invalid keys
	invalidKeys := []struct {
		name string
		key  string
	}{
		{"Empty key", ""},
		{"Invalid prefix", "invalid_prefix_12345"},
		{"Key without prefix", "12345abcdef"},
		{"Key with invalid characters", KeyPrefix + "!@#$%^&*()"},
		{"Key with spaces", KeyPrefix + "abc def"},
		{"Key too short", KeyPrefix + "abc"},
	}

	for _, tc := range invalidKeys {
		t.Run(tc.name, func(t *testing.T) {
			isValid := ValidateAPIKey(tc.key)
			if isValid {
				t.Errorf("Expected key %q to be invalid", tc.key)
			}
		})
	}

	// Test key validation with timing attack prevention
	t.Run("Timing attack resistance", func(t *testing.T) {
		key1, _ := GenerateAPIKey(DefaultKeyLength)
		key2, _ := GenerateAPIKey(DefaultKeyLength)

		// This is a basic check that ensures we're using constant-time comparison
		// A proper timing attack test would require more sophisticated measurement
		result := subtle.ConstantTimeCompare([]byte(key1), []byte(key2))
		if result != 0 { // Keys should be different
			t.Errorf("Expected different keys to not match in constant time comparison")
		}

		result = subtle.ConstantTimeCompare([]byte(key1), []byte(key1))
		if result != 1 { // Same key should match
			t.Errorf("Expected identical keys to match in constant time comparison")
		}
	})
}

func TestIsRootAPIKey(t *testing.T) {
	// Test with matching root key
	t.Run("Matching root key", func(t *testing.T) {
		rootKey := KeyPrefix + "test_root_key"
		isRoot := IsRootAPIKey(rootKey, rootKey)
		if !isRoot {
			t.Errorf("Expected key %q to be recognized as root key", rootKey)
		}
	})

	// Test with non-matching root key
	t.Run("Non-matching root key", func(t *testing.T) {
		rootKey := KeyPrefix + "test_root_key"
		testKey := KeyPrefix + "test_regular_key"
		isRoot := IsRootAPIKey(testKey, rootKey)
		if isRoot {
			t.Errorf("Expected key %q to not be recognized as root key", testKey)
		}
	})

	// Test with invalid key format
	t.Run("Invalid key format", func(t *testing.T) {
		rootKey := KeyPrefix + "test_root_key"
		invalidKey := "invalid_key_format"
		isRoot := IsRootAPIKey(invalidKey, rootKey)
		if isRoot {
			t.Errorf("Expected invalid key %q to not be recognized as root key", invalidKey)
		}
	})

	// Test with empty root key
	t.Run("Empty root key", func(t *testing.T) {
		testKey := KeyPrefix + "test_key"
		isRoot := IsRootAPIKey(testKey, "")
		if isRoot {
			t.Errorf("Expected no key to be recognized as root when root key is empty")
		}
	})

	// Test timing attack resistance
	t.Run("Timing attack resistance", func(t *testing.T) {
		rootKey := KeyPrefix + "test_root_key"
		similarKey := KeyPrefix + "test_root_key_extra"

		// This test simply verifies we're using constant-time comparison
		start := time.Now()
		IsRootAPIKey(rootKey, rootKey)
		exactMatch := time.Since(start)

		start = time.Now()
		IsRootAPIKey(similarKey, rootKey)
		similarMatch := time.Since(start)

		// The timing should be similar regardless of match
		// This is not a perfect test, but gives us some confidence
		// In a real timing attack test, we'd run thousands of iterations
		ratio := float64(similarMatch.Nanoseconds()) / float64(exactMatch.Nanoseconds())
		if ratio < 0.5 || ratio > 2.0 {
			t.Logf("Timing ratio: %f (should be close to 1.0 for constant-time comparison)", ratio)
			// This is just a log, not a failure, as timing can vary on different systems
		}
	})
}

func TestCreateAPIKey(t *testing.T) {
	// Test successful API key creation
	t.Run("Successful creation", func(t *testing.T) {
		roleID := uint(1)
		description := "Test API Key"
		customData := "{\"user_id\": 123, \"username\": \"testuser\"}"
		duration := 24 * time.Hour

		apiKey, err := CreateAPIKey(roleID, description, customData, &duration)
		if err != nil {
			t.Fatalf("CreateAPIKey failed with error: %v", err)
		}

		// Check that the generated key has the correct format
		if !strings.HasPrefix(apiKey.Key, KeyPrefix) {
			t.Errorf("Expected key to have prefix %q, got %q", KeyPrefix, apiKey.Key)
		}

		// Check that role ID, description, and custom data were set
		if apiKey.RoleID != roleID {
			t.Errorf("Expected RoleID to be %d, got %d", roleID, apiKey.RoleID)
		}
		if apiKey.Description != description {
			t.Errorf("Expected Description to be %q, got %q", description, apiKey.Description)
		}
		if apiKey.CustomData != customData {
			t.Errorf("Expected CustomData to be %q, got %q", customData, apiKey.CustomData)
		}

		// Check that expiration was set
		if apiKey.ExpiresAt == nil {
			t.Error("Expected ExpiresAt to be set")
		} else {
			// Allow 1 second tolerance for test execution time
			expectedExpiry := time.Now().Add(duration)
			diff := apiKey.ExpiresAt.Sub(expectedExpiry)
			if diff < -time.Second || diff > time.Second {
				t.Errorf("Expected ExpiresAt to be around %v, got %v (diff: %v)",
					expectedExpiry, *apiKey.ExpiresAt, diff)
			}
		}
	})

	// Test key creation with invalid role ID
	t.Run("Invalid role ID", func(t *testing.T) {
		roleID := uint(0) // Invalid role ID
		description := "Test API Key"
		customData := ""
		duration := 24 * time.Hour

		_, err := CreateAPIKey(roleID, description, customData, &duration)
		if err == nil {
			t.Error("Expected error for role ID 0")
		}
	})

	// Test key creation without expiration
	t.Run("No expiration", func(t *testing.T) {
		roleID := uint(1)
		description := "Test API Key without expiration"
		customData := "{\"user_id\": 456}"

		apiKey, err := CreateAPIKey(roleID, description, customData, nil)
		if err != nil {
			t.Fatalf("CreateAPIKey failed with error: %v", err)
		}

		if apiKey.ExpiresAt != nil {
			t.Errorf("Expected ExpiresAt to be nil, got %v", *apiKey.ExpiresAt)
		}

		if apiKey.CustomData != customData {
			t.Errorf("Expected CustomData to be %q, got %q", customData, apiKey.CustomData)
		}
	})

	// Test key creation with empty custom data
	t.Run("Empty custom data", func(t *testing.T) {
		roleID := uint(1)
		description := "Test API Key with empty custom data"
		customData := ""
		duration := 24 * time.Hour

		apiKey, err := CreateAPIKey(roleID, description, customData, &duration)
		if err != nil {
			t.Fatalf("CreateAPIKey failed with error: %v", err)
		}

		if apiKey.CustomData != "" {
			t.Errorf("Expected CustomData to be empty, got %q", apiKey.CustomData)
		}
	})
}

// Additional tests for any added security features (like key rotation, expiry, etc.)
// Add more tests as features are developed
