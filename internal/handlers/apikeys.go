package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/BasementPilot/orbit-keys/internal/database"
	"github.com/BasementPilot/orbit-keys/internal/models"
	"github.com/BasementPilot/orbit-keys/utils"
	"gorm.io/gorm"
)

// CreateAPIKeyRequest defines the request structure for creating a new API key.
// It specifies the role to associate with the key, an optional description,
// and an optional expiration time in days.
type CreateAPIKeyRequest struct {
	RoleID      uint   `json:"role_id" validate:"required"`
	Description string `json:"description"`
	ExpiresIn   *int   `json:"expires_in"` // Expiration in days, nil means no expiration
}

// CreateAPIKey handles requests to create a new API key.
// It validates the request, checks that the specified role exists,
// generates a secure API key, and stores it in the database.
//
// Returns:
// - 201 Created with the created API key on success
// - 400 Bad Request if the request body is invalid or the role doesn't exist
// - 500 Internal Server Error if a database or key generation error occurs
func CreateAPIKey(c *fiber.Ctx) error {
	// Parse request body
	req := new(CreateAPIKeyRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.RoleID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Role ID is required",
		})
	}

	// Verify role exists
	db := database.GetDB()
	var role models.Role
	if err := db.First(&role, req.RoleID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	// Convert expiration days to time.Duration
	var expiresIn *time.Duration
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		days := time.Duration(*req.ExpiresIn) * 24 * time.Hour
		expiresIn = &days
	}

	// Create API key
	apiKey, err := utils.CreateAPIKey(req.RoleID, req.Description, expiresIn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate API key",
		})
	}

	// Save to database
	if err := db.Create(apiKey).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save API key",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(apiKey)
}

// GetAPIKeys handles requests to retrieve all API keys from the database.
// It returns a JSON array of all API keys with their associated roles preloaded.
//
// Returns:
// - 200 OK with an array of API keys on success
// - 500 Internal Server Error if a database error occurs
func GetAPIKeys(c *fiber.Ctx) error {
	var apiKeys []models.APIKey
	db := database.GetDB()
	
	// Get all API keys with their roles
	if err := db.Preload("Role").Find(&apiKeys).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve API keys",
		})
	}

	return c.JSON(apiKeys)
}

// GetAPIKey handles requests to retrieve a single API key by its ID.
// The ID is extracted from the URL parameters, and the role is preloaded.
//
// Returns:
// - 200 OK with the requested API key on success
// - 404 Not Found if the API key doesn't exist
func GetAPIKey(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var apiKey models.APIKey
	db := database.GetDB()
	
	// Find API key by ID with its role
	if err := db.Preload("Role").Where("id = ?", id).First(&apiKey).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	return c.JSON(apiKey)
}

// DeleteAPIKey handles requests to delete an API key by its ID.
// It first checks if the API key exists before deleting it.
//
// Returns:
// - 204 No Content on successful deletion
// - 404 Not Found if the API key doesn't exist
// - 500 Internal Server Error if a database error occurs
func DeleteAPIKey(c *fiber.Ctx) error {
	id := c.Params("id")
	
	db := database.GetDB()
	
	// Find API key by ID to ensure it exists
	var apiKey models.APIKey
	if err := db.First(&apiKey, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	// Delete the API key
	if err := db.Delete(&apiKey).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete API key",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// LookupAPIKey finds and validates an API key by its key string.
// This endpoint is typically used internally to verify a key during authentication.
// It also updates the last used timestamp for the key when found.
//
// Returns:
// - 200 OK with the API key details on success
// - 400 Bad Request if the key parameter is missing
// - 401 Unauthorized if the key has expired
// - 404 Not Found if the key doesn't exist
func LookupAPIKey(c *fiber.Ctx) error {
	// This endpoint is for internal use only, protected by the root API key
	key := c.Query("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Key parameter is required",
		})
	}

	var apiKey models.APIKey
	db := database.GetDB()
	
	// Find API key by key value with its role
	if err := db.Preload("Role").Where("key = ?", key).First(&apiKey).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	// Check if expired
	if apiKey.IsExpired() {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "API key has expired",
			"key":   apiKey,
		})
	}

	// Update last used timestamp
	go func(db *gorm.DB, apiKey *models.APIKey) {
		if err := apiKey.UpdateLastUsed(db); err != nil {
			// Log the error but don't block the request
			log.Printf("Failed to update LastUsed timestamp: %v", err)
		}
	}(db, &apiKey)

	return c.JSON(apiKey)
}

// ValidateAPIKeyPermission checks if an API key has a specific permission.
// This endpoint is typically used internally during authorization checks.
// It also updates the last used timestamp for the key when queried.
//
// Returns:
// - 200 OK with a boolean indicating if the key has the permission
// - 400 Bad Request if key or permission parameters are missing
// - 401 Unauthorized if the key has expired
// - 404 Not Found if the key doesn't exist
func ValidateAPIKeyPermission(c *fiber.Ctx) error {
	// This endpoint is for internal use only, protected by the root API key
	key := c.Query("key")
	permission := c.Query("permission")
	
	if key == "" || permission == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Key and permission parameters are required",
		})
	}

	var apiKey models.APIKey
	db := database.GetDB()
	
	// Find API key by key value with its role
	if err := db.Preload("Role").Where("key = ?", key).First(&apiKey).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	// Check if expired
	if apiKey.IsExpired() {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "API key has expired",
		})
	}

	// Check if the key has the required permission
	hasPermission := apiKey.Role.HasPermission(permission)
	
	// Update last used timestamp
	go func(db *gorm.DB, apiKey *models.APIKey) {
		if err := apiKey.UpdateLastUsed(db); err != nil {
			// Log the error but don't block the request
			log.Printf("Failed to update LastUsed timestamp: %v", err)
		}
	}(db, &apiKey)

	return c.JSON(fiber.Map{
		"has_permission": hasPermission,
	})
}

// UpdateAPIKeyExpiration updates the expiration date of an existing API key.
// It can set a new expiration, remove the expiration, or expire the key immediately.
//
// Returns:
// - 200 OK with the updated API key on success
// - 400 Bad Request if the ID is invalid or the request body is malformed
// - 404 Not Found if the API key doesn't exist
// - 500 Internal Server Error if a database error occurs
func UpdateAPIKeyExpiration(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid API key ID",
		})
	}
	
	// Parse request body
	type ExpirationRequest struct {
		ExpiresIn *int `json:"expires_in"` // Days from now, nil to remove expiration
	}
	
	req := new(ExpirationRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Find API key
	db := database.GetDB()
	var apiKey models.APIKey
	if err := db.First(&apiKey, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	// Update expiration
	if req.ExpiresIn == nil {
		// Remove expiration
		apiKey.ExpiresAt = nil
	} else if *req.ExpiresIn < 0 {
		// Expire immediately
		now := time.Now()
		apiKey.ExpiresAt = &now
	} else {
		// Set new expiration
		expiresAt := time.Now().Add(time.Duration(*req.ExpiresIn) * 24 * time.Hour)
		apiKey.ExpiresAt = &expiresAt
	}

	// Save changes
	if err := db.Save(&apiKey).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update API key expiration",
		})
	}

	return c.JSON(apiKey)
} 