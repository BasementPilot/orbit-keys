# OrbitKeys - API Key Management for Go Fiber

OrbitKeys is a comprehensive API key management system for Go applications using the Fiber web framework. It provides a complete solution for secure API key generation, role-based permissions, and middleware integration with robust security features.

## Why OrbitKeys Exists

We built OrbitKeys while struggling with the same challenges many Go developers face when implementing authentication systems. After wrestling with API key management in several projects, it became clear that this common need shouldn't require everyone to reinvent the wheel.

### The Problems We're Trying to Help With

- **Authentication Complexity**: Building proper authentication from scratch is time-consuming and easy to get wrong
- **Security Concerns**: Implementing all the security best practices takes significant research and experience
- **Permission Headaches**: As applications grow, managing who can do what quickly becomes complicated
- **Maintenance Burden**: Keeping authentication systems updated and secure requires ongoing attention

OrbitKeys is simply an attempt to share a solution to these common challenges. It's designed to help developers focus on building their unique application features instead of spending time on authentication infrastructure that's needed in almost every API project.

## Features

- **Secure API Key Management**
  - Generate cryptographically secure API keys
  - Protection against timing attacks and brute force attempts
  - Assign roles with specific permissions to API keys
  - Set expiration dates for API keys
  - Track when keys were last used

- **Role-Based Permissions**
  - Create roles with specific permission sets
  - Permission format: `resource:action` (e.g., "users:read")
  - Wildcard support: `resource:*` or `*` for all permissions
  - Prevent deletion of roles in use by API keys

- **Security-Focused Middleware**
  - Rate limiting for authentication attempts
  - Timeout protection against long-running requests
  - Validate API keys and permissions with secure comparisons
  - Generic error messages to prevent information disclosure
  - Update "last used" timestamp on API key usage

- **Admin API**
  - Root API key for administrative operations
  - Full CRUD operations for API keys and roles
  - Utility endpoints for validating permissions

## Installation

```bash
go get github.com/BasementPilot/orbit-keys
```

## Configuration

OrbitKeys uses environment variables for configuration:

- `ORBITKEYS_ROOT_API_KEY`: The master API key for admin access (generated if not provided)
- `ORBITKEYS_DB_PATH`: Path to the SQLite database file (default: "orbitkeys.db")
- `ORBITKEYS_BASE_URL`: Base URL for API endpoints (default: "/api")

You can use a `.env` file or set these variables in your environment. The SQLite database will be created automatically if it doesn't exist.

## Quick Start

```go
package main

import (
    "log"
    
    "github.com/gofiber/fiber/v2"
    "github.com/BasementPilot/orbit-keys"
)

func main() {
    // Initialize OrbitKeys with default configuration
    ok, err := orbitkeys.New(nil)
    if err != nil {
        log.Fatalf("Failed to initialize OrbitKeys: %v", err)
    }
    
    // Set up the service
    if err := ok.Init(); err != nil {
        log.Fatalf("Failed to initialize service: %v", err)
    }
    
    // Start the server - this blocks until shutdown
    if err := ok.Start(":3000"); err != nil {
        log.Fatalf("Server error: %v", err)
    }
    
    // Graceful shutdown
    defer ok.Shutdown()
}
```

## Custom Application Integration

If you want to integrate OrbitKeys with your existing Fiber application:

```go
package main

import (
    "log"
    
    "github.com/gofiber/fiber/v2"
    "github.com/BasementPilot/orbit-keys"
    "github.com/BasementPilot/orbit-keys/internal/middleware"
    "github.com/BasementPilot/orbit-keys/config"
)

func main() {
    // Create your Fiber app
    app := fiber.New()
    
    // Initialize OrbitKeys configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // Create OrbitKeys instance with custom config
    ok, err := orbitkeys.New(cfg)
    if err != nil {
        log.Fatalf("Failed to initialize OrbitKeys: %v", err)
    }
    
    // Initialize the service
    if err := ok.Init(); err != nil {
        log.Fatalf("Failed to initialize service: %v", err)
    }
    
    // Add your routes
    api := app.Group("/api")
    
    // Protected route requiring "users:read" permission
    users := api.Group("/users")
    users.Use(middleware.APIKeyAuth("users:read"))
    users.Get("/", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "message": "This is a protected endpoint",
            "users":   []string{"user1", "user2", "user3"},
        })
    })
    
    // Start your server
    app.Listen(":3000")
    
    // Graceful shutdown
    defer ok.Shutdown()
}
```

## API Authentication

To authenticate API requests, include the API key in the header:

```
X-API-Key: orbitkey_your_api_key_here
```

For admin operations, use the root API key:

```
X-Root-API-Key: orbitkey_root_your_root_api_key_here
```

## API Endpoints

### API Key Validation Endpoints (protected by root API key)

- `GET /api/lookup?key=xxx` - Look up API key details
- `GET /api/validate?key=xxx&permission=xxx` - Check if key has permission

### Role Management Endpoints

- `GET /api/roles` - List all roles (requires "roles:read" permission)
- `GET /api/roles/:id` - Get role details (requires "roles:read" permission)
- `POST /api/roles` - Create a new role (requires "roles:create" permission)
- `PUT /api/roles/:id` - Update a role (requires "roles:update" permission)
- `DELETE /api/roles/:id` - Delete a role (requires "roles:delete" permission)

### API Key Management Endpoints

- `GET /api/keys` - List all API keys (requires "keys:read" permission)
- `GET /api/keys/:id` - Get API key details (requires "keys:read" permission)
- `POST /api/keys` - Create a new API key (requires "keys:create" permission)
- `PUT /api/keys/:id/expiration` - Update API key expiration (requires "keys:update" permission)
- `DELETE /api/keys/:id` - Delete an API key (requires "keys:delete" permission)

## Creating Custom Middleware

You can create custom middleware for your application:

```go
// Basic API key authentication without permission check
app.Use(middleware.APIKeyAuth(""))

// Require specific permission
app.Post("/users", middleware.APIKeyAuth("users:write"), createUserHandler)

// Rate limiting
app.Use(middleware.CreateRateLimiter(100, 1*time.Minute))

// Multiple middleware layers
adminRoute := app.Group("/admin")
adminRoute.Use(middleware.APIKeyAuth("")) // Just authenticate the API key
adminRoute.Use(middleware.RequirePermission("admin:access")) // Then check permissions
```

## Example Requests

### Creating a Role

```bash
curl -X POST http://localhost:3000/api/roles \
  -H "X-API-Key: orbitkey_your_api_key_with_roles_create_permission" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "basic-user",
    "description": "Basic user with read-only access",
    "permissions": ["users:read", "products:read"]
  }'
```

### Creating an API Key

```bash
curl -X POST http://localhost:3000/api/keys \
  -H "X-API-Key: orbitkey_your_api_key_with_keys_create_permission" \
  -H "Content-Type: application/json" \
  -d '{
    "role_id": 1,
    "description": "API key for Example App",
    "expires_in": 30
  }'
```

### Using an API Key

```bash
curl http://localhost:3000/api/protected-resource \
  -H "X-API-Key: orbitkey_your_api_key"
```

## Security Considerations

- The system implements rate limiting to prevent brute force attacks
- API keys are validated using constant-time comparisons to prevent timing attacks
- Environment variables are sanitized to prevent injection attacks
- File paths are validated to prevent directory traversal attacks
- Generic error messages are used to prevent information disclosure
- Timeouts are implemented to prevent long-running requests

## License

MIT 