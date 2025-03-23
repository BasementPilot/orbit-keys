package orbitkeys

import (
	"testing"

	"github.com/BasementPilot/orbit-keys/config"
)

func TestNew(t *testing.T) {
	// Test with nil config (should load default)
	t.Run("Nil config", func(t *testing.T) {
		ok, err := New(nil)
		if err != nil {
			t.Fatalf("New() with nil config failed: %v", err)
		}
		if ok == nil {
			t.Fatal("New() returned nil OrbitKeys")
		}
		if ok.Config == nil {
			t.Fatal("Config is nil after initialization with nil config")
		}
		if ok.Config.RootAPIKey == "" {
			t.Error("RootAPIKey is empty after initialization")
		}
		if ok.Config.DBPath != "orbitkeys.db" {
			t.Errorf("Expected default DBPath 'orbitkeys.db', got '%s'", ok.Config.DBPath)
		}
		if ok.Config.BaseURL != "/api" {
			t.Errorf("Expected default BaseURL '/api', got '%s'", ok.Config.BaseURL)
		}
	})

	// Test with custom config
	t.Run("Custom config", func(t *testing.T) {
		cfg := &config.Config{
			RootAPIKey: "orbitkey_test_root_key",
			DBPath:     "test.db",
			BaseURL:    "/custom",
		}

		ok, err := New(cfg)
		if err != nil {
			t.Fatalf("New() with custom config failed: %v", err)
		}
		if ok == nil {
			t.Fatal("New() returned nil OrbitKeys")
		}
		if ok.Config.RootAPIKey != cfg.RootAPIKey {
			t.Errorf("Expected RootAPIKey '%s', got '%s'", cfg.RootAPIKey, ok.Config.RootAPIKey)
		}
		if ok.Config.DBPath != cfg.DBPath {
			t.Errorf("Expected DBPath '%s', got '%s'", cfg.DBPath, ok.Config.DBPath)
		}
		if ok.Config.BaseURL != cfg.BaseURL {
			t.Errorf("Expected BaseURL '%s', got '%s'", cfg.BaseURL, ok.Config.BaseURL)
		}
	})
}

func TestInit(t *testing.T) {
	// Skip tests that require database connectivity
	t.Skip("Skipping tests that require database connectivity")
}

func TestStartAndShutdown(t *testing.T) {
	// Skip tests that require database connectivity
	t.Skip("Skipping tests that require database connectivity")
}

// Placeholder for future integration tests of API endpoints
// These would use the httptest package to create requests and verify responses
// For now, we've focused on unit testing the core components 