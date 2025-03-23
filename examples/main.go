package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	orbitkeys "github.com/BasementPilot/orbit-keys"
	"github.com/BasementPilot/orbit-keys/config"
	"github.com/BasementPilot/orbit-keys/internal/middleware"
	"github.com/BasementPilot/orbit-keys/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize OrbitKeys with config
	ok, err := orbitkeys.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize OrbitKeys: %v", err)
	}

	// Initialize OrbitKeys service
	if err := ok.Init(); err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	// Ensure proper shutdown when done
	defer ok.Shutdown()

	// Create our own Fiber app for custom routes
	app := fiber.New()

	// Add basic middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Powered-By", "OrbitKeys Example")
		return c.Next()
	})

	// Add rate limiting for all routes
	app.Use(middleware.CreateRateLimiter(100, 1*time.Minute))

	// Example protected routes
	api := app.Group("/api")

	// Protected routes with different permission requirements
	users := api.Group("/users")
	users.Use(middleware.APIKeyAuth("users:read")) // Require users:read permission
	users.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "This is a protected users endpoint",
			"users":   []string{"user1", "user2", "user3"},
		})
	})

	products := api.Group("/products")
	products.Use(middleware.APIKeyAuth("products:read")) // Require products:read permission
	products.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":  "This is a protected products endpoint",
			"products": []string{"product1", "product2", "product3"},
		})
	})

	// Example endpoint using custom data from API key
	profile := api.Group("/profile")
	profile.Use(middleware.APIKeyAuth("profile:read")) // Require profile:read permission
	profile.Get("/", func(c *fiber.Ctx) error {
		// Get the API key from the context
		apiKey, ok := c.Locals("apiKey").(models.APIKey)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "API key not found in context",
			})
		}

		// Parse the custom data
		customData, err := apiKey.GetCustomData()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse custom data",
			})
		}

		// Check if we have user information in the custom data
		if customData == nil {
			return c.JSON(fiber.Map{
				"message": "No user data found in API key",
			})
		}

		// Get user information from custom data
		userID, hasUserID := customData["user_id"]
		username, hasUsername := customData["username"]

		if !hasUserID || !hasUsername {
			return c.JSON(fiber.Map{
				"message": "Incomplete user data in API key",
				"data":    customData,
			})
		}

		// Use the custom data to personalize the response
		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("Hello, %v! Your user ID is %v.", username, userID),
			"user": fiber.Map{
				"id":       userID,
				"username": username,
				"data":     customData,
			},
		})
	})

	// Admin route with multiple permissions (checked after authentication)
	admin := api.Group("/admin")
	admin.Use(middleware.APIKeyAuth(""))               // Authenticate API key without checking permissions yet
	admin.Use(middleware.RequirePermission("admin:*")) // Then check for admin:* permission
	admin.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "This is a protected admin endpoint",
		})
	})

	// Public routes that don't require API key
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to the OrbitKeys example application",
		})
	})

	// Start the server in a goroutine
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}

		log.Printf("Server starting on port %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}
	log.Println("Server gracefully stopped")
}
