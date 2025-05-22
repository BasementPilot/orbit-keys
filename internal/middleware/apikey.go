// Package middleware provides Fiber middleware for API key authentication and authorization.
// It handles validating API keys, checking permissions, and enforcing security policies
// for protected routes in the OrbitKeys system.
package middleware

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/BasementPilot/orbit-keys/config"
	"github.com/BasementPilot/orbit-keys/internal/database"
	"github.com/BasementPilot/orbit-keys/internal/models"
	"github.com/BasementPilot/orbit-keys/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"gorm.io/gorm"
)

// APIKeyHeader defines the HTTP header name used for API key authentication.
// Clients must include this header with a valid API key to access protected routes.
const APIKeyHeader = "X-API-Key"

// RootAPIKeyHeader defines the HTTP header name used for root API key authentication.
// This header is used for administrative operations that require elevated privileges.
const RootAPIKeyHeader = "X-Root-API-Key"

// authAttempts tracks failed authentication attempts by IP address
var (
	authAttempts     = make(map[string]int)
	authAttemptsMux  sync.RWMutex
	attemptThreshold = 10 // Max failed attempts before rate limiting
)

// APIKeyAuth creates middleware that authenticates and authorizes requests using API keys.
// It verifies the API key exists in the header, validates its format, checks if it exists
// in the database, ensures it hasn't expired, and verifies it has the required permission.
//
// The requiredPermission parameter specifies what permission is needed to access the route.
// If empty, it only verifies the API key is valid without checking permissions.
//
// When authentication succeeds, the API key and role are stored in the request context
// for use by subsequent handlers.
func APIKeyAuth(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get(APIKeyHeader)

		// Check for rate limiting if client has too many failed attempts
		ip := c.IP()
		authAttemptsMux.RLock()
		attempts, exists := authAttempts[ip]
		authAttemptsMux.RUnlock()

		if exists && attempts >= attemptThreshold {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many failed authentication attempts, please try again later",
			})
		}

		if apiKey == "" {
			// Track failed authentication attempt
			trackFailedAttempt(ip)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "API key is required",
			})
		}

		// Check if it's a valid API key format
		if !utils.ValidateAPIKey(apiKey) {
			// Track failed authentication attempt
			trackFailedAttempt(ip)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key format",
			})
		}

		// Find the API key in the database
		var key models.APIKey
		db := database.GetDB()
		if err := db.Preload("Role").Where("key = ?", apiKey).First(&key).Error; err != nil {
			// Track failed authentication attempt
			trackFailedAttempt(ip)

			// Use generic error message to avoid information disclosure
			if jsonErr := c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication failed",
			}); jsonErr != nil {
				log.Printf("Error sending JSON response: %v", jsonErr)
				return jsonErr
			}
			return nil
		}

		// Check if the API key has expired
		if key.IsExpired() {
			// We don't track this as a failed attempt since it's a valid key
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "API key has expired",
			})
		}

		// Check if the API key has the required permission
		if requiredPermission != "" && !key.Role.HasPermission(requiredPermission) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
			})
		}

		// Reset failed attempt counter on successful authentication
		if exists {
			authAttemptsMux.Lock()
			delete(authAttempts, ip)
			authAttemptsMux.Unlock()
		}

		// Update the last used timestamp asynchronously
		go func(db *gorm.DB, key *models.APIKey) {
			if err := key.UpdateLastUsed(db); err != nil {
				log.Printf("Failed to update LastUsed timestamp: %v", err)
			}
		}(db, &key)

		// Store API key and role information in context for later use
		c.Locals("apiKey", key)
		c.Locals("role", key.Role)

		return c.Next()
	}
}

// RootAPIKeyAuth creates middleware that authenticates requests using the root API key.
// This middleware is used to protect administrative endpoints that require elevated privileges.
// It checks if the root API key header is present and matches the configured root API key.
//
// The cfg parameter provides the configuration containing the root API key to check against.
func RootAPIKeyAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rootKey := c.Get(RootAPIKeyHeader)

		// Check for rate limiting if client has too many failed attempts
		ip := c.IP()
		authAttemptsMux.RLock()
		attempts, exists := authAttempts[ip]
		authAttemptsMux.RUnlock()

		if exists && attempts >= attemptThreshold {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many failed authentication attempts, please try again later",
			})
		}

		if rootKey == "" {
			// Track failed authentication attempt
			trackFailedAttempt(ip)

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Root API key is required for admin operations",
			})
		}

		// Check if it matches the configured root API key
		if !utils.IsRootAPIKey(rootKey, cfg.RootAPIKey) {
			// Track failed authentication attempt
			trackFailedAttempt(ip)

			// Use generic error message to avoid information disclosure
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication failed",
			})
		}

		// Reset failed attempt counter on successful authentication
		if exists {
			authAttemptsMux.Lock()
			delete(authAttempts, ip)
			authAttemptsMux.Unlock()
		}

		return c.Next()
	}
}

// RequirePermission creates middleware that checks if an authenticated API key has a specific permission.
// This middleware should be used after the APIKeyAuth middleware, as it relies on the role
// being stored in the request context.
//
// The permission parameter specifies what permission is needed to access the route.
func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the role from context
		role, ok := c.Locals("role").(models.Role)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Check if the role has the required permission
		if !role.HasPermission(permission) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": fmt.Sprintf("API key does not have the required permission: %s", permission),
			})
		}

		return c.Next()
	}
}

// CreateRateLimiter returns a middleware for rate limiting requests based on IP address.
// It helps protect against brute force attacks and DoS attempts.
//
// Parameters:
//   - max: Maximum number of requests allowed in the time window
//   - expiration: Duration of the time window
//
// Returns a configured rate limiter middleware.
func CreateRateLimiter(max int, expiration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Rate limit by IP address
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded, please try again later",
			})
		},
	})
}

// trackFailedAttempt increments the failed authentication attempts counter for an IP address.
// This is used to implement progressive rate limiting for potential brute force attacks.
func trackFailedAttempt(ip string) {
	authAttemptsMux.Lock()
	defer authAttemptsMux.Unlock()

	authAttempts[ip]++

	// Clean up old attempts periodically to prevent memory leaks
	// In production, this should be handled by a dedicated goroutine or cache with TTL
}
