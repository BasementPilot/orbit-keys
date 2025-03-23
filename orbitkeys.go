package orbitkeys

import (
	"crypto/rand"
	"encoding/base64"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/BasementPilot/orbit-keys/config"
	"github.com/BasementPilot/orbit-keys/internal/database"
	"github.com/BasementPilot/orbit-keys/internal/handlers"
	"github.com/BasementPilot/orbit-keys/internal/middleware"
)

// OrbitKeys represents the API key management system
type OrbitKeys struct {
	Config *config.Config
	App    *fiber.App
}

// New creates a new instance of OrbitKeys
func New() (*OrbitKeys, error) {
	// Load configuration
	cfg := config.LoadConfig()

	// Generate root API key if none is provided
	if cfg.RootAPIKey == "" {
		key, err := generateRootAPIKey()
		if err != nil {
			return nil, err
		}
		cfg.RootAPIKey = key
		
		if err := config.SaveConfig(cfg); err != nil {
			log.Printf("Warning: Failed to save root API key to .env file: %v", err)
		} else {
			log.Printf("Generated new root API key: %s", key)
			log.Println("This key has been saved to the .env file")
		}
	}

	// Initialize database
	if err := database.InitDB(cfg); err != nil {
		return nil, err
	}

	// Create default admin role
	if err := database.CreateDefaultAdminRole(); err != nil {
		return nil, err
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Create OrbitKeys instance
	orbitKeys := &OrbitKeys{
		Config: cfg,
		App:    app,
	}

	// Setup routes
	orbitKeys.setupRoutes()

	return orbitKeys, nil
}

// setupRoutes sets up all the API routes
func (o *OrbitKeys) setupRoutes() {
	baseURL := o.Config.BaseURL
	if baseURL == "" {
		baseURL = "/api"
	}

	// API Group
	api := o.App.Group(baseURL)

	// Admin routes - protected by root API key
	admin := api.Group("/admin")
	admin.Use(middleware.RootAPIKeyAuth(o.Config))

	// API Key Management
	admin.Post("/api-keys", handlers.CreateAPIKey)
	admin.Get("/api-keys", handlers.GetAPIKeys)
	admin.Get("/api-keys/:id", handlers.GetAPIKey)
	admin.Delete("/api-keys/:id", handlers.DeleteAPIKey)
	admin.Patch("/api-keys/:id/expiration", handlers.UpdateAPIKeyExpiration)

	// Role Management
	admin.Post("/roles", handlers.CreateRole)
	admin.Get("/roles", handlers.GetRoles)
	admin.Get("/roles/:id", handlers.GetRole)
	admin.Put("/roles/:id", handlers.UpdateRole)
	admin.Delete("/roles/:id", handlers.DeleteRole)

	// Utility API key endpoints - protected by root API key
	admin.Get("/lookup-key", handlers.LookupAPIKey)
	admin.Get("/validate-permission", handlers.ValidateAPIKeyPermission)

	// Public API health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})
}

// GetMiddleware returns middleware for validating API keys with the required permission
func (o *OrbitKeys) GetMiddleware(permission string) fiber.Handler {
	return middleware.APIKeyAuth(permission)
}

// RequirePermission returns middleware to check if the authenticated API key has a specific permission
func (o *OrbitKeys) RequirePermission(permission string) fiber.Handler {
	return middleware.RequirePermission(permission)
}

// generateRootAPIKey generates a new cryptographically secure root API key
func generateRootAPIKey() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	
	return "orbitkey_root_" + base64.URLEncoding.EncodeToString(bytes), nil
}

// Close properly closes the OrbitKeys system
func (o *OrbitKeys) Close() {
	database.CloseDB()
} 