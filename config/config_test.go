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
	origRootAPIKey := os.Getenv("ORBITKEYS_ROOT_API_KEY")
	origDBPath := os.Getenv("ORBITKEYS_DB_PATH")
	origBaseURL := os.Getenv("ORBITKEYS_BASE_URL")
	defer func() {
		os.Setenv("ORBITKEYS_ROOT_API_KEY", origRootAPIKey)
		os.Setenv("ORBITKEYS_DB_PATH", origDBPath)
		os.Setenv("ORBITKEYS_BASE_URL", origBaseURL)
	}()

	// Test with valid environment variables
	t.Run("Valid environment variables", func(t *testing.T) {
		os.Setenv("ORBITKEYS_ROOT_API_KEY", "test_root_key")
		os.Setenv("ORBITKEYS_DB_PATH", "test.db")
		os.Setenv("ORBITKEYS_BASE_URL", "api")

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
	})

	// Test with invalid path
	t.Run("Invalid DB path", func(t *testing.T) {
		os.Setenv("ORBITKEYS_ROOT_API_KEY", "test_root_key")
		os.Setenv("ORBITKEYS_DB_PATH", "../test.db") // Directory traversal
		os.Setenv("ORBITKEYS_BASE_URL", "api")

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
		os.Setenv("ORBITKEYS_ROOT_API_KEY", "test_root_key")
		os.Setenv("ORBITKEYS_DB_PATH", "test.db")
		os.Setenv("ORBITKEYS_BASE_URL", "api/")

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
		os.Unsetenv("ORBITKEYS_ROOT_API_KEY")
		os.Unsetenv("ORBITKEYS_DB_PATH")
		os.Unsetenv("ORBITKEYS_BASE_URL")

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
	})
}

func TestSaveConfig(t *testing.T) {
	// Create a test config
	config := &Config{
		RootAPIKey: "test_root_key",
		DBPath:     "test.db",
		BaseURL:    "/api",
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
		// Rename existing .env file if any
		if _, err := os.Stat(".env"); err == nil {
			err = os.Rename(".env", ".env.backup")
			if err != nil {
				t.Fatalf("Failed to rename existing .env file: %v", err)
			}
			defer os.Rename(".env.backup", ".env")
		}

		// Delete temporary files if they exist
		os.Remove(".env.tmp")
		os.Remove(".env")
		defer os.Remove(".env") // Clean up after test

		err := SaveConfig(config)
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
		}

		for _, line := range expectedLines {
			if !contains(content, line) {
				t.Errorf("Expected .env file to contain %q", line)
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
			},
			expected: false,
		},
		{
			name: "Valid config",
			config: &Config{
				RootAPIKey: "test_root_key",
				DBPath:     "test.db",
				BaseURL:    "/api",
			},
			expected: true,
		},
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

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
} 