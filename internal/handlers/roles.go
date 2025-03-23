// Package handlers provides HTTP request handlers for the OrbitKeys API.
// It contains implementations for all API endpoints, handling the request parsing,
// validation, business logic, and response formatting.
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/BasementPilot/orbit-keys/internal/database"
	"github.com/BasementPilot/orbit-keys/internal/models"
)

// CreateRoleRequest defines the request structure for creating a new role.
// It contains the required name, optional description, and a list of permissions.
type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions" validate:"required"`
}

// UpdateRoleRequest defines the request structure for updating an existing role.
// All fields are optional, as updates can be partial.
type UpdateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// CreateRole handles role creation requests.
// It validates the request body, ensures permissions are in the correct format,
// and creates a new role in the database.
//
// Returns:
// - 201 Created with the created role on success
// - 400 Bad Request if the request body is invalid or contains invalid permissions
// - 500 Internal Server Error if a database error occurs
func CreateRole(c *fiber.Ctx) error {
	// Parse request body
	req := new(CreateRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	// Convert permissions array to string
	permissions := ""
	for i, p := range req.Permissions {
		if !models.ValidatePermissionFormat(p) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid permission format: " + p,
			})
		}
		
		if i > 0 {
			permissions += ","
		}
		permissions += p
	}

	// Create role
	role := models.Role{
		Name:        req.Name,
		Description: req.Description,
		Permissions: permissions,
	}

	// Save to database
	db := database.GetDB()
	if err := db.Create(&role).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create role: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(role)
}

// GetRoles handles requests to retrieve all roles from the database.
// It returns a JSON array of all roles with their associated data.
//
// Returns:
// - 200 OK with an array of roles on success
// - 500 Internal Server Error if a database error occurs
func GetRoles(c *fiber.Ctx) error {
	var roles []models.Role
	db := database.GetDB()
	
	// Get all roles
	if err := db.Find(&roles).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve roles",
		})
	}

	return c.JSON(roles)
}

// GetRole handles requests to retrieve a single role by its ID.
// The ID is extracted from the URL parameters.
//
// Returns:
// - 200 OK with the requested role on success
// - 404 Not Found if the role doesn't exist
func GetRole(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var role models.Role
	db := database.GetDB()
	
	// Find role by ID
	if err := db.Where("id = ?", id).First(&role).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	return c.JSON(role)
}

// UpdateRole handles requests to update an existing role.
// It retrieves the role by ID, updates the provided fields,
// and saves the changes to the database.
//
// Returns:
// - 200 OK with the updated role on success
// - 400 Bad Request if the request body is invalid or contains invalid permissions
// - 404 Not Found if the role doesn't exist
// - 500 Internal Server Error if a database error occurs
func UpdateRole(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Parse request body
	req := new(UpdateRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Find role by ID
	db := database.GetDB()
	var role models.Role
	if err := db.Where("id = ?", id).First(&role).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	// Update fields if provided
	if req.Name != "" {
		role.Name = req.Name
	}
	
	if req.Description != "" {
		role.Description = req.Description
	}
	
	if req.Permissions != nil {
		// Convert permissions array to string
		permissions := ""
		for i, p := range req.Permissions {
			if !models.ValidatePermissionFormat(p) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid permission format: " + p,
				})
			}
			
			if i > 0 {
				permissions += ","
			}
			permissions += p
		}
		role.Permissions = permissions
	}

	// Save to database
	if err := db.Save(&role).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update role: " + err.Error(),
		})
	}

	return c.JSON(role)
}

// DeleteRole handles requests to delete a role by its ID.
// It first checks if any API keys are using this role, preventing
// deletion if the role is in use to maintain data integrity.
//
// Returns:
// - 204 No Content on successful deletion
// - 400 Bad Request if the role is assigned to API keys
// - 500 Internal Server Error if a database error occurs
func DeleteRole(c *fiber.Ctx) error {
	id := c.Params("id")
	
	db := database.GetDB()
	
	// Check if the role is in use by any API keys
	var count int64
	if err := db.Model(&models.APIKey{}).Where("role_id = ?", id).Count(&count).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check role usage",
		})
	}
	
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot delete role as it is assigned to API keys",
		})
	}

	// Delete the role
	if err := db.Delete(&models.Role{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete role",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
} 