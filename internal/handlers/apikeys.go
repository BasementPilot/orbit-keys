package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/BasementPilot/orbit-keys/internal/database"
	"github.com/BasementPilot/orbit-keys/internal/models"
	"github.com/BasementPilot/orbit-keys/utils"
)

// API key request and response structures
type CreateAPIKeyRequest struct {
	RoleID      uint   `json:"role_id" validate:"required"`
	Description string `json:"description"`
	ExpiresIn   *int   `json:"expires_in"` // Expiration in days, nil means no expiration
}

// CreateAPIKey handles API key creation
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

// GetAPIKeys handles retrieving all API keys
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

// GetAPIKey handles retrieving a single API key by ID
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

// DeleteAPIKey handles deleting an API key
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

// LookupAPIKey finds an API key by its key string
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
	go apiKey.UpdateLastUsed(db)

	return c.JSON(apiKey)
}

// ValidateAPIKeyPermission validates if an API key has a specific permission
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
	go apiKey.UpdateLastUsed(db)

	return c.JSON(fiber.Map{
		"has_permission": hasPermission,
	})
}

// UpdateAPIKeyExpiration updates an API key's expiration date
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