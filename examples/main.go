package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	orbitkeys "github.com/BasementPilot/orbit-keys"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize OrbitKeys
	ok, err := orbitkeys.New()
	if err != nil {
		log.Fatalf("Failed to initialize OrbitKeys: %v", err)
	}
	defer ok.Close()

	// Create a Fiber app with OrbitKeys
	app := ok.App // Use the app from OrbitKeys or create your own

	// Example protected routes
	api := app.Group("/api")

	// Protected routes with different permission requirements
	users := api.Group("/users")
	users.Use(ok.GetMiddleware("users:read")) // Require users:read permission
	users.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "This is a protected users endpoint",
			"users":   []string{"user1", "user2", "user3"},
		})
	})

	products := api.Group("/products")
	products.Use(ok.GetMiddleware("products:read")) // Require products:read permission
	products.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":  "This is a protected products endpoint",
			"products": []string{"product1", "product2", "product3"},
		})
	})

	// Admin route with multiple permissions (checked after authentication)
	admin := api.Group("/admin")
	admin.Use(ok.GetMiddleware("")) // Authenticate API key without checking permissions yet
	admin.Use(ok.RequirePermission("admin:*")) // Then check for admin:* permission
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