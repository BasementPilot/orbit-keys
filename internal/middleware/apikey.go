package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/BasementPilot/orbit-keys/config"
	"github.com/BasementPilot/orbit-keys/internal/database"
	"github.com/BasementPilot/orbit-keys/internal/models"
	"github.com/BasementPilot/orbit-keys/utils"
)

// APIKeyHeader is the name of the header that contains the API key
const APIKeyHeader = "X-API-Key"

// RootAPIKeyHeader is the name of the header that contains the root API key
const RootAPIKeyHeader = "X-Root-API-Key"

// APIKeyAuth middleware checks for a valid API key
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

// RootAPIKeyAuth middleware checks for a valid root API key
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

// RequirePermission middleware checks if the authenticated user has the required permission
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