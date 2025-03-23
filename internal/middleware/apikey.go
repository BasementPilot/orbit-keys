// Package middleware provides Fiber middleware for API key authentication and authorization.
// It handles validating API keys, checking permissions, and enforcing security policies
// for protected routes in the OrbitKeys system.
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/BasementPilot/orbit-keys/config"
	"github.com/BasementPilot/orbit-keys/internal/database"
	"github.com/BasementPilot/orbit-keys/internal/models"
	"github.com/BasementPilot/orbit-keys/utils"
)

// APIKeyHeader defines the HTTP header name used for API key authentication.
// Clients must include this header with a valid API key to access protected routes.
const APIKeyHeader = "X-API-Key"

// RootAPIKeyHeader defines the HTTP header name used for root API key authentication.
// This header is used for administrative operations that require elevated privileges.
const RootAPIKeyHeader = "X-Root-API-Key"

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
		// Get the API key from the header
		apiKey := c.Get(APIKeyHeader)
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "API key is required",
			})
		}

		// Check if it's a valid API key format
		if !utils.ValidateAPIKey(apiKey) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key format",
			})
		}

		// Find the API key in the database
		var key models.APIKey
		db := database.GetDB()
		if err := db.Preload("Role").Where("key = ?", apiKey).First(&key).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key",
			})
		}

		// Check if the API key has expired
		if key.IsExpired() {
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

		// Update the last used timestamp
		go key.UpdateLastUsed(db)

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
		// Get the root API key from the header
		rootKey := c.Get(RootAPIKeyHeader)
		if rootKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Root API key is required for admin operations",
			})
		}

		// Check if it matches the configured root API key
		if !utils.IsRootAPIKey(rootKey, cfg.RootAPIKey) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid root API key",
			})
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
				"error": "Insufficient permissions",
			})
		}

		return c.Next()
	}
} 