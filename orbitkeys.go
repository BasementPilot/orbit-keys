// Package orbitkeys provides API key management functionality for applications.
// It offers secure generation, validation, and permission-based authorization
// of API keys through a REST API and programmatic interfaces.
package orbitkeys

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/BasementPilot/orbit-keys/config"
	"github.com/BasementPilot/orbit-keys/internal/database"
	"github.com/BasementPilot/orbit-keys/internal/handlers"
	"github.com/BasementPilot/orbit-keys/internal/middleware"
	"github.com/BasementPilot/orbit-keys/utils"
)

// OrbitKeys represents the API key management service.
// It contains the configuration and manages the lifecycle of the service.
type OrbitKeys struct {
	Config *config.Config
	app    *fiber.App
}

// New creates a new instance of the OrbitKeys service with the provided configuration.
// If no configuration is provided, default values will be used.
// It initializes the service but does not start the server.
func New(cfg *config.Config) (*OrbitKeys, error) {
	// If no config provided, load default
	var err error
	if cfg == nil {
		cfg, err = config.LoadConfig()
		if err != nil {
			return nil, err
		}
	}

	// Validate configuration
	if !config.ValidateConfig(cfg) {
		// Generate root API key if not provided
		rootKey, err := utils.GenerateAPIKey(utils.DefaultKeyLength)
		if err != nil {
			return nil, err
		}
		
		cfg.RootAPIKey = rootKey
		log.Printf("Generated new root API key: %s", rootKey)
		
		// Save configuration
		if err := config.SaveConfig(cfg); err != nil {
			log.Printf("Warning: Failed to save configuration: %v", err)
		}
	}

	return &OrbitKeys{
		Config: cfg,
	}, nil
}

// Init initializes the OrbitKeys service components including the database and API routes.
// It must be called before Start().
func (o *OrbitKeys) Init() error {
	// Initialize database
	if err := database.InitDB(o.Config); err != nil {
		return err
	}

	// Create default admin role if it doesn't exist
	if err := database.CreateDefaultAdminRole(); err != nil {
		return err
	}

	// Create Fiber app with middleware
	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		// Explicitly set ErrorHandler to customize error responses
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			
			// Don't expose internal error details in production
			errMsg := "Internal Server Error"
			if code != fiber.StatusInternalServerError {
				errMsg = err.Error()
			}
			
			return c.Status(code).JSON(fiber.Map{
				"error": errMsg,
			})
		},
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())
	
	// Add rate limiting for all routes
	app.Use(middleware.CreateRateLimiter(100, 1*time.Minute))

	// Setup API routes
	apiGroup := app.Group(o.Config.BaseURL)

	// Public endpoints (authenticated with root API key)
	publicGroup := apiGroup.Group("")
	publicGroup.Use(middleware.RootAPIKeyAuth(o.Config))
	
	// Add additional rate limiting for authentication endpoints
	authRateLimiter := middleware.CreateRateLimiter(30, 5*time.Minute)
	publicGroup.Use(authRateLimiter)

	// API key lookup and validation endpoints
	publicGroup.Get("/lookup", handlers.LookupAPIKey)
	publicGroup.Get("/validate", handlers.ValidateAPIKeyPermission)

	// Role management endpoints
	roleGroup := apiGroup.Group("/roles")
	roleGroup.Use(middleware.APIKeyAuth("roles:read"))
	roleGroup.Get("/", handlers.GetRoles)
	roleGroup.Get("/:id", handlers.GetRole)
	
	// Protected by stronger permissions
	roleGroup.Post("/", middleware.RequirePermission("roles:create"), handlers.CreateRole)
	roleGroup.Put("/:id", middleware.RequirePermission("roles:update"), handlers.UpdateRole)
	roleGroup.Delete("/:id", middleware.RequirePermission("roles:delete"), handlers.DeleteRole)

	// API key management endpoints
	keyGroup := apiGroup.Group("/keys")
	keyGroup.Use(middleware.APIKeyAuth("keys:read"))
	keyGroup.Get("/", handlers.GetAPIKeys)
	keyGroup.Get("/:id", handlers.GetAPIKey)
	
	// Protected by stronger permissions
	keyGroup.Post("/", middleware.RequirePermission("keys:create"), handlers.CreateAPIKey)
	keyGroup.Put("/:id/expiration", middleware.RequirePermission("keys:update"), handlers.UpdateAPIKeyExpiration)
	keyGroup.Delete("/:id", middleware.RequirePermission("keys:delete"), handlers.DeleteAPIKey)

	o.app = app
	return nil
}

// Start runs the HTTP server and starts accepting requests.
// Init() must be called before this method.
func (o *OrbitKeys) Start(address string) error {
	if o.app == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Server not initialized. Call Init() first")
	}
	
	log.Printf("OrbitKeys service starting on %s", address)
	log.Printf("Root API Key: %s", o.Config.RootAPIKey)
	return o.app.Listen(address)
}

// Shutdown gracefully stops the server and closes database connections.
func (o *OrbitKeys) Shutdown() error {
	if o.app != nil {
		if err := o.app.Shutdown(); err != nil {
			return err
		}
	}
	
	database.CloseDB()
	log.Println("OrbitKeys service shutdown complete")
	return nil
} 
} 