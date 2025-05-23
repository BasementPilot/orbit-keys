package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BasementPilot/orbit-keys/config"
	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
)

func setupTestApp(cfg *config.Config) (*fiber.App, error) {
	if cfg == nil {
		// Provide a default config if nil is passed, to avoid nil pointer dereferences
		// and ensure tests not focused on config still run.
		cfg = &config.Config{
			RootAPIKey: "orbitkey_test_root_key", // Default for RootAPIKeyAuth
			AuthTimeoutSeconds: 2, // Default for APIKeyAuth
		}
	}
	if cfg.RootAPIKey == "" { // Ensure RootAPIKey is set for RootAPIKeyAuth
		cfg.RootAPIKey = "orbitkey_test_root_key"
	}
	if cfg.AuthTimeoutSeconds == 0 { // Ensure AuthTimeoutSeconds is non-zero
		cfg.AuthTimeoutSeconds = 2
	}


	app := fiber.New()

	// Initialize test routes with middleware
	app.Get("/protected", APIKeyAuth(cfg, "test:permission"), func(c *fiber.Ctx) error {
		return c.SendString("Protected content")
	})

	app.Get("/root-only", RootAPIKeyAuth(cfg), func(c *fiber.Ctx) error {
		// This handler now correctly runs after RootAPIKeyAuth
		return c.SendString("Root only content")
	})

	return app, nil
}

func TestAPIKeyAuth(t *testing.T) {
	// Skip this test if we can't set up the database
	t.Skip("Skipping API key auth test as it requires database setup")

	// Set up test app
	app, err := setupTestApp(nil) // Pass nil to use default config
	if err != nil {
		t.Fatalf("Failed to set up test app: %v", err)
	}
	
	// Test cases
	tests := []struct {
		name       string
		apiKey     string
		statusCode int
	}{
		{
			name:       "No API key",
			apiKey:     "",
			statusCode: fiber.StatusUnauthorized,
		},
		{
			name:       "Invalid API key format",
			apiKey:     "invalid-key-format",
			statusCode: fiber.StatusUnauthorized,
		},
		{
			name:       "Valid key format but not in DB",
			apiKey:     "orbitkey_nonexistent_key",
			statusCode: fiber.StatusUnauthorized,
		},
		// Add more test cases if you can set up a test database
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new http request
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.apiKey != "" {
				req.Header.Set(APIKeyHeader, tc.apiKey)
			}
			
			// Perform the request
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test failed: %v", err)
			}
			
			if resp.StatusCode != tc.statusCode {
				t.Errorf("Expected status code %d, got %d", tc.statusCode, resp.StatusCode)
			}
		})
	}
}

func TestRootAPIKeyAuth(t *testing.T) {
	// Set up test app with mock functionality
	app, err := setupTestApp(nil) // Pass nil to use default config
	if err != nil {
		t.Fatalf("Failed to set up test app: %v", err)
	}
	
	// Test cases
	tests := []struct {
		name       string
		apiKey     string
		statusCode int
	}{
		{
			name:       "No API key",
			apiKey:     "",
			statusCode: fiber.StatusUnauthorized,
		},
		{
			name:       "Invalid API key format",
			apiKey:     "invalid-key-format",
			statusCode: fiber.StatusUnauthorized,
		},
		{
			name:       "Valid key format but not root",
			apiKey:     "orbitkey_not_root_key",
			statusCode: fiber.StatusUnauthorized,
		},
		{
			name:       "Valid root key",
			apiKey:     "orbitkey_test_root_key",
			statusCode: fiber.StatusOK,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new http request
			req := httptest.NewRequest(http.MethodGet, "/root-only", nil)
			if tc.apiKey != "" {
				req.Header.Set(RootAPIKeyHeader, tc.apiKey)
			}
			
			// Perform the request
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test failed: %v", err)
			}
			
			if resp.StatusCode != tc.statusCode {
				t.Errorf("Expected status code %d, got %d", tc.statusCode, resp.StatusCode)
			}
		})
	}
}

func TestRequirePermission(t *testing.T) {
	// We'll test this without using the actual fiber context since we're 
	// just mocking the Role in Locals() which is challenging in tests
	
	// Skip this test as it requires proper mocking of Fiber context
	t.Skip("Skipping RequirePermission test as it requires proper Fiber context mocking")
}

func TestCreateRateLimiter(t *testing.T) {
	// Set up test app with rate limiter
	app := fiber.New()
	
	// Use a very low limit to test rate limiting easily
	app.Use(CreateRateLimiter(2, 1*time.Second))
	
	app.Get("/rate-limited", func(c *fiber.Ctx) error {
		return c.SendString("Limited content")
	})
	
	// Make multiple requests in a short time to trigger rate limiting
	req := httptest.NewRequest(http.MethodGet, "/rate-limited", nil)
	
	// First request - should succeed
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status code %d for first request, got %d", fiber.StatusOK, resp.StatusCode)
	}
	
	// Second request - should succeed
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status code %d for second request, got %d", fiber.StatusOK, resp.StatusCode)
	}
	
	// Third request - should be rate limited
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Third request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Errorf("Expected status code %d for third request, got %d", fiber.StatusTooManyRequests, resp.StatusCode)
	}
} 

func TestAPIKeyAuth_RateLimitingWithTTLCache(t *testing.T) {
	originalAuthCache := authCache
	originalAttemptThreshold := attemptThreshold

	// Configure a short TTL and small threshold for testing
	authCache = cache.New(2*time.Second, 1*time.Second)
	attemptThreshold = 2

	defer func() {
		authCache = originalAuthCache
		attemptThreshold = originalAttemptThreshold
	}()

	app, err := setupTestApp(nil) // Use default config for this test
	if err != nil {
		t.Fatalf("Failed to set up test app: %v", err)
	}

	// Simulate failed attempts
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(APIKeyHeader, "orbitkey_nonexistent_for_ratelimit_test") // Invalid key
	req.RemoteAddr = "1.2.3.4:12345" // Mock IP address

	// First attempt (threshold = 2)
	resp, _ := app.Test(req, -1) // -1 for no timeout on app.Test
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d for 1st failed attempt, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}

	// Second attempt - should now be rate limited
	resp, _ = app.Test(req, -1)
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Errorf("Expected status %d for 2nd failed attempt (rate limited), got %d", fiber.StatusTooManyRequests, resp.StatusCode)
	}

	// Wait for cache to expire (TTL is 2s, wait for 3s)
	time.Sleep(3 * time.Second)

	// Third attempt - should no longer be rate limited, but still unauthorized
	resp, _ = app.Test(req, -1)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d after TTL expiry, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestAPIKeyAuth_Timeout(t *testing.T) {
	// This test attempts to verify the timeout mechanism.
	// Actual timeout duration accuracy (e.g., exactly 2 seconds) is hard to test reliably in automated unit tests
	// and is typically verified by configuration and manual/integration testing.
	// This test uses a very short timeout to check if the timeout logic path can be triggered.

	testCfg := &config.Config{
		AuthTimeoutSeconds: 1, // 1 second timeout
		RootAPIKey:         "orbitkey_test_root_key", // Needed for setupTestApp
	}
	app, err := setupTestApp(testCfg)
	if err != nil {
		t.Fatalf("Failed to set up test app: %v", err)
	}

	// To reliably trigger the timeout, the authentication process (including DB lookup)
	// would need to consistently take longer than AuthTimeoutSeconds.
	// For this unit test, we set a very short timeout (1s).
	// A real DB lookup for a non-existent key is usually very fast.
	// If this test becomes flaky, it means the DB call + overhead is faster than 1s.
	// A more robust test would involve mocking the DB to introduce a guaranteed delay,
	// which is beyond the current scope of changes.

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(APIKeyHeader, "orbitkey_somekey_for_timeout_test") // Non-existent key

	// We expect the request to timeout because AuthTimeoutSeconds is 1.
	// The internal goroutine in APIKeyAuth doing the DB lookup might not finish
	// before the `time.After` in the select statement triggers.
	resp, err := app.Test(req, -1) // Using -1 for app.Test timeout, relying on middleware timeout
	if err != nil {
		// If app.Test itself times out (which it shouldn't if -1 is used correctly),
		// or another error occurs.
		t.Fatalf("app.Test failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusRequestTimeout {
		// This might happen if the auth process (including DB check) completes faster than AuthTimeoutSeconds.
		// For a 1-second timeout, this is possible.
		t.Logf("Auth process completed faster than the 1s timeout. Received status: %d. DB lookup might be too fast.", resp.StatusCode)
		t.Errorf("Expected status code %d (StatusRequestTimeout), got %d", fiber.StatusRequestTimeout, resp.StatusCode)
	}

	// It's also worth noting that the Fiber test framework's own timeout (`app.Test(req, timeout)`)
	// can interfere if not set appropriately (e.g. to -1 for no timeout, or a value
	// significantly larger than the middleware timeout being tested).
}