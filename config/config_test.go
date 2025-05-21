package config

import (
	"os"
	"strings"
	"testing"
)

func TestSanitizeEnv(t *testing.T) {
	testCases := []struct {
		name          string
		value         string
		shouldSanitize bool
		expected      string
	}{
		{
			name:          "Valid value",
			value:         "test-value",
			shouldSanitize: false,
			expected:      "test-value",
		},
		{
			name:          "Value with spaces",
			value:         "  test-value  ",
			shouldSanitize: false,
			expected:      "test-value",
		},
		{
			name:          "Value with semicolon (dangerous)",
			value:         "test;value",
			shouldSanitize: true,
			expected:      "",
		},
		{
			name:          "Value with pipe (dangerous)",
			value:         "test|value",
			shouldSanitize: true,
			expected:      "",
		},
		{
			name:          "Value with backtick (dangerous)",
			value:         "test`value",
			shouldSanitize: true,
			expected:      "",
		},
		{
			name:          "Value with dollar sign (dangerous)",
			value:         "test$value",
			shouldSanitize: true,
			expected:      "",
		},
		{
			name:          "Value with less than sign (dangerous)",
			value:         "test<value",
			shouldSanitize: true,
			expected:      "",
		},
		{
			name:          "Value with greater than sign (dangerous)",
			value:         "test>value",
			shouldSanitize: true,
			expected:      "",
		},
		{
			name:          "Value with parentheses (dangerous)",
			value:         "test()value",
			shouldSanitize: true,
			expected:      "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment variable
			testKey := "TEST_ORBITKEYS_VAR"
			err := os.Setenv(testKey, tc.value)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv(testKey)

			// Call the function under test
			result := sanitizeEnv(testKey)

			// Check the result
			if result != tc.expected {
				t.Errorf("Expected sanitizeEnv to return %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestIsValidFilePath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Valid path",
			path:     "data/test.db",
			expected: true,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "Path with directory traversal",
			path:     "../test.db",
			expected: false,
		},
		{
			name:     "Path with multiple directory traversal",
			path:     "data/../../test.db",
			expected: false,
		},
		{
			name:     "Absolute path",
			path:     "/tmp/test.db",
			expected: true,
		},
		{
			name:     "Current directory path",
			path:     "./test.db",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidFilePath(tc.path)
			if result != tc.expected {
				t.Errorf("Expected isValidFilePath(%q) to return %v, got %v", tc.path, tc.expected, result)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Back up original environment variables
	// Back up original environment variables
	// Using t.Setenv for these will handle cleanup automatically if we modify them.
	// However, LoadConfig reads multiple, so we'll manage them explicitly here for clarity if needed later.
	origEnv := map[string]string{
		"ORBITKEYS_ROOT_API_KEY":     os.Getenv("ORBITKEYS_ROOT_API_KEY"),
		"ORBITKEYS_DB_PATH":          os.Getenv("ORBITKEYS_DB_PATH"),
		"ORBITKEYS_BASE_URL":         os.Getenv("ORBITKEYS_BASE_URL"),
		"ORBITKEYS_AUTH_TIMEOUT_SECONDS": os.Getenv("ORBITKEYS_AUTH_TIMEOUT_SECONDS"),
	}
	defer func() {
		for key, val := range origEnv {
			if val == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, val)
			}
		}
	}()

	// Test with valid environment variables
	t.Run("Valid environment variables", func(t *testing.T) {
		t.Setenv("ORBITKEYS_ROOT_API_KEY", "test_root_key")
		t.Setenv("ORBITKEYS_DB_PATH", "test.db")
		t.Setenv("ORBITKEYS_BASE_URL", "api")
		t.Setenv("ORBITKEYS_AUTH_TIMEOUT_SECONDS", "5")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig returned an error: %v", err)
		}

		if config.RootAPIKey != "test_root_key" {
			t.Errorf("Expected RootAPIKey to be 'test_root_key', got %q", config.RootAPIKey)
		}
		if config.DBPath != "test.db" {
			t.Errorf("Expected DBPath to be 'test.db', got %q", config.DBPath)
		}
		if config.BaseURL != "/api" {
			t.Errorf("Expected BaseURL to be '/api', got %q", config.BaseURL)
		}
		if config.AuthTimeoutSeconds != 5 {
			t.Errorf("Expected AuthTimeoutSeconds to be 5, got %d", config.AuthTimeoutSeconds)
		}
	})

	// Test with invalid path
	t.Run("Invalid DB path", func(t *testing.T) {
		t.Setenv("ORBITKEYS_ROOT_API_KEY", "test_root_key")
		t.Setenv("ORBITKEYS_DB_PATH", "../test.db") // Directory traversal
		t.Setenv("ORBITKEYS_BASE_URL", "api")
		t.Setenv("ORBITKEYS_AUTH_TIMEOUT_SECONDS", "2")


		config, err := LoadConfig()
		if err == nil {
			t.Fatal("LoadConfig should have returned an error for invalid path")
		}
		if err != ErrInvalidFilePath {
			t.Errorf("Expected ErrInvalidFilePath, got %v", err)
		}

		// Even with error, config should be populated
		if config.RootAPIKey != "test_root_key" {
			t.Errorf("Expected RootAPIKey to be 'test_root_key', got %q", config.RootAPIKey)
		}
	})

	// Test with path normalization (trailing slash)
	t.Run("BaseURL normalization", func(t *testing.T) {
		t.Setenv("ORBITKEYS_ROOT_API_KEY", "test_root_key")
		t.Setenv("ORBITKEYS_DB_PATH", "test.db")
		t.Setenv("ORBITKEYS_BASE_URL", "api/")
		t.Setenv("ORBITKEYS_AUTH_TIMEOUT_SECONDS", "2")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig returned an error: %v", err)
		}

		if config.BaseURL != "/api" {
			t.Errorf("Expected BaseURL to be '/api', got %q", config.BaseURL)
		}
	})

	// Test with empty variables (using defaults)
	t.Run("Empty variables (defaults)", func(t *testing.T) {
		// Unset specific variables for this test
		// t.Setenv would require setting them to empty string, os.Unsetenv is clearer here.
		currentRootKey := os.Getenv("ORBITKEYS_ROOT_API_KEY")
		currentDBPath := os.Getenv("ORBITKEYS_DB_PATH")
		currentBaseURL := os.Getenv("ORBITKEYS_BASE_URL")
		currentAuthTimeout := os.Getenv("ORBITKEYS_AUTH_TIMEOUT_SECONDS")

		os.Unsetenv("ORBITKEYS_ROOT_API_KEY")
		os.Unsetenv("ORBITKEYS_DB_PATH")
		os.Unsetenv("ORBITKEYS_BASE_URL")
		os.Unsetenv("ORBITKEYS_AUTH_TIMEOUT_SECONDS")

		defer func() { // Restore them after this sub-test
			os.Setenv("ORBITKEYS_ROOT_API_KEY", currentRootKey)
			os.Setenv("ORBITKEYS_DB_PATH", currentDBPath)
			os.Setenv("ORBITKEYS_BASE_URL", currentBaseURL)
			os.Setenv("ORBITKEYS_AUTH_TIMEOUT_SECONDS", currentAuthTimeout)
		}()

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig returned an error: %v", err)
		}

		if config.DBPath != "orbitkeys.db" {
			t.Errorf("Expected default DBPath to be 'orbitkeys.db', got %q", config.DBPath)
		}
		if config.BaseURL != "/api" {
			t.Errorf("Expected default BaseURL to be '/api', got %q", config.BaseURL)
		}
		if config.AuthTimeoutSeconds != 2 { // Default value for AuthTimeoutSeconds
			t.Errorf("Expected default AuthTimeoutSeconds to be 2, got %d", config.AuthTimeoutSeconds)
		}
	})

	// Test LoadConfig with invalid AuthTimeoutSeconds
	t.Run("Invalid AuthTimeoutSeconds", func(t *testing.T) {
		t.Setenv("ORBITKEYS_ROOT_API_KEY", "test_root_key")
		t.Setenv("ORBITKEYS_DB_PATH", "test.db")
		t.Setenv("ORBITKEYS_BASE_URL", "/api")
		t.Setenv("ORBITKEYS_AUTH_TIMEOUT_SECONDS", "not-an-int")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig returned an error: %v", err)
		}
		if config.AuthTimeoutSeconds != 2 { // Should use default
			t.Errorf("Expected AuthTimeoutSeconds to default to 2 for invalid input, got %d", config.AuthTimeoutSeconds)
		}
	})
}

func TestSaveConfig(t *testing.T) {
	// Create a test config
	testConf := &Config{ // Renamed to avoid conflict
		RootAPIKey:         "test_root_key",
		DBPath:             "test.db",
		BaseURL:            "/api",
		AuthTimeoutSeconds: 5, // Include the new field
	}

	// Test with nil config
	t.Run("Nil config", func(t *testing.T) {
		err := SaveConfig(nil)
		if err == nil {
			t.Fatal("SaveConfig should return an error for nil config")
		}
	})

	// Test actual save to a temporary file
	t.Run("Save to file", func(t *testing.T) {
		// Manage .env file for testing
		const envFile = ".env"
		const backupEnvFile = ".env.backup"

		// Back up existing .env if it exists
		if _, err := os.Stat(envFile); err == nil {
			if err := os.Rename(envFile, backupEnvFile); err != nil {
				t.Fatalf("Failed to rename existing %s to %s: %v", envFile, backupEnvFile, err)
			}
			defer func() { // Ensure backup is restored
				if err := os.Rename(backupEnvFile, envFile); err != nil {
					// If the test created a new .env, backup might not exist.
					// This is a best-effort restore.
					if !os.IsNotExist(err) {
						t.Logf("Warning: Failed to restore %s from %s: %v", envFile, backupEnvFile, err)
					}
				}
			}()
		} else if !os.IsNotExist(err) {
			t.Fatalf("Failed to stat %s: %v", envFile, err) // Real error other than not existing
		}
		
		// Clean up any .env file created by the test
		defer os.Remove(envFile)
		// Clean up .env.tmp as well, if SaveConfig fails before rename
		defer os.Remove(".env.tmp")


		err := SaveConfig(testConf) // Use the renamed testConf
		if err != nil {
			t.Fatalf("SaveConfig returned an error: %v", err)
		}

		// Verify the file was created
		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			t.Fatal("SaveConfig did not create .env file")
		}

		// Read the file contents
		data, err := os.ReadFile(".env")
		if err != nil {
			t.Fatalf("Failed to read .env file: %v", err)
		}

		// Check file contents
		content := string(data)
		expectedLines := []string{
			"ORBITKEYS_ROOT_API_KEY=test_root_key",
			"ORBITKEYS_DB_PATH=test.db",
			"ORBITKEYS_BASE_URL=/api",
			"ORBITKEYS_AUTH_TIMEOUT_SECONDS=5", // Check for the new field
		}

		for _, line := range expectedLines {
			if !strings.Contains(content, line) { // Using strings.Contains directly
				t.Errorf("Expected .env file to contain %q. Full content:\n%s", line, content)
			}
		}
	})
}

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name:     "Nil config",
			config:   nil,
			expected: false,
		},
		{
			name: "Empty RootAPIKey",
			config: &Config{
				RootAPIKey: "",
				DBPath:     "test.db",
				BaseURL:    "/api",
				AuthTimeoutSeconds: 2,
			},
			expected: false,
		},
		{
			name: "Valid config",
			config: &Config{
				RootAPIKey: "test_root_key",
				DBPath:     "test.db",
				BaseURL:    "/api",
				AuthTimeoutSeconds: 2,
			},
			expected: true,
		},
		// AuthTimeoutSeconds being 0 or negative is not explicitly a validation error by ValidateConfig,
		// as LoadConfig assigns a default. If it were a critical validation point,
		// ValidateConfig would need to check it.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateConfig(tc.config)
			if result != tc.expected {
				t.Errorf("Expected ValidateConfig to return %v, got %v", tc.expected, result)
			}
		})
	}
}


func TestParseEnvToInt(t *testing.T) {
	testCases := []struct {
		name         string
		envKey       string
		envValue     string // Value to set for the env var
		shouldSetEnv bool   // Whether to set the env var or leave it unset
		defaultValue int
		expectedValue int
	}{
		{
			name:          "Valid integer string",
			envKey:        "TEST_PARSE_INT_VALID",
			envValue:      "123",
			shouldSetEnv:  true,
			defaultValue:  0,
			expectedValue: 123,
		},
		{
			name:          "Invalid integer string",
			envKey:        "TEST_PARSE_INT_INVALID",
			envValue:      "abc",
			shouldSetEnv:  true,
			defaultValue:  42,
			expectedValue: 42, // Should return defaultValue
		},
		{
			name:          "Empty string (variable not set)",
			envKey:        "TEST_PARSE_INT_EMPTY",
			shouldSetEnv:  false, // Do not set the env var
			defaultValue:  77,
			expectedValue: 77, // Should return defaultValue
		},
		{
			name:          "Environment variable set to empty string",
			envKey:        "TEST_PARSE_INT_SET_EMPTY",
			envValue:      "", // Explicitly set to empty
			shouldSetEnv:  true,
			defaultValue:  88,
			expectedValue: 88, // Should return defaultValue
		},
		{
			name:          "Negative integer string",
			envKey:        "TEST_PARSE_INT_NEGATIVE",
			envValue:      "-5",
			shouldSetEnv:  true,
			defaultValue:  0,
			expectedValue: -5,
		},
		{
			name:          "Zero string",
			envKey:        "TEST_PARSE_INT_ZERO",
			envValue:      "0",
			shouldSetEnv:  true,
			defaultValue:  10,
			expectedValue: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldSetEnv {
				t.Setenv(tc.envKey, tc.envValue)
			} else {
				// Ensure it's unset if previously set by another test or system env
				// t.Setenv(tc.envKey, "") followed by os.Unsetenv(tc.envKey) would also work
				// but this is cleaner if we know it should be considered "unset".
				// For this test structure, simply not calling t.Setenv is sufficient
 spezifischeally if the keys are unique like TEST_PARSE_INT_EMPTY.
				// However, to be absolutely sure it's not lingering from a prior non-test setenv:
				originalValue, isSet := os.LookupEnv(tc.envKey)
				if isSet {
					os.Unsetenv(tc.envKey)
					defer os.Setenv(tc.envKey, originalValue)
				}
			}

			result := parseEnvToInt(tc.envKey, tc.defaultValue)
			if result != tc.expectedValue {
				t.Errorf("parseEnvToInt(%q, %d) with env value %q: expected %d, got %d",
					tc.envKey, tc.defaultValue, tc.envValue, tc.expectedValue, result)
			}
		})
	}
} 

// Helper function to check if a string contains a substring - REMOVED as direct strings.Contains is fine
// func contains(s, substr string) bool {
// 	return strings.Contains(s, substr)
// }