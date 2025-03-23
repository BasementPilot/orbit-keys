# OrbitKeys - API Key Management for Go Fiber

OrbitKeys is a comprehensive API key management system for Go applications using the Fiber web framework. It provides a complete solution for API key generation, role-based permissions, and middleware integration.

## Features

- **API Key Management**
  - Generate cryptographically secure API keys
  - Assign roles with specific permissions to API keys
  - Set expiration dates for API keys
  - Track when keys were last used

- **Role-Based Permissions**
  - Create roles with specific permission sets
  - Permission format: `resource:action` (e.g., "users:read")
  - Wildcard support: `resource:*` or `*` for all permissions
  - Prevent deletion of roles in use by API keys

- **Middleware Integration**
  - Easy integration with Fiber routes
  - Validate API keys and permissions
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

You can use a `.env` file or set these variables in your environment.

## Quick Start

```go
package main

import (
    "log"
    
    "github.com/gofiber/fiber/v2"
    "github.com/BasementPilot/orbit-keys"
)

func main() {
    // Initialize OrbitKeys
    ok, err := orbitkeys.New()
    if err != nil {
        log.Fatalf("Failed to initialize OrbitKeys: %v", err)
    }
    defer ok.Close()
    
    // Use the app from OrbitKeys
    app := ok.App
    
    // Create protected routes
    api := app.Group("/api")
    
    // Protected route requiring "users:read" permission
    users := api.Group("/users")
    users.Use(ok.GetMiddleware("users:read"))
    users.Get("/", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "message": "This is a protected endpoint",
            "users":   []string{"user1", "user2", "user3"},
        })
    })
    
    // Start the server
    app.Listen(":3000")
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

### Admin API Endpoints (protected by root API key)

#### API Key Management

- `POST /api/admin/api-keys` - Create a new API key
- `GET /api/admin/api-keys` - List all API keys
- `GET /api/admin/api-keys/:id` - Get API key details
- `DELETE /api/admin/api-keys/:id` - Delete an API key
- `PATCH /api/admin/api-keys/:id/expiration` - Update API key expiration

#### Role Management

- `POST /api/admin/roles` - Create a new role
- `GET /api/admin/roles` - List all roles
- `GET /api/admin/roles/:id` - Get role details
- `PUT /api/admin/roles/:id` - Update a role
- `DELETE /api/admin/roles/:id` - Delete a role

#### Utility Endpoints

- `GET /api/admin/lookup-key?key=xxx` - Look up API key details
- `GET /api/admin/validate-permission?key=xxx&permission=xxx` - Check if key has permission

### Public Endpoint

- `GET /api/health` - Health check endpoint

## Creating Custom Middleware

You can create custom middleware with specific permission requirements:

```go
// Require "users:write" permission
app.Post("/api/users", ok.GetMiddleware("users:write"), createUserHandler)

// Multiple permission checks
adminRoute := app.Group("/api/admin")
adminRoute.Use(ok.GetMiddleware("")) // Authenticate without checking permissions yet
adminRoute.Use(ok.RequirePermission("admin:access")) // Then check for admin access
```

## Example Requests

### Creating a Role

```bash
curl -X POST http://localhost:3000/api/admin/roles \
  -H "X-Root-API-Key: orbitkey_root_your_root_key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "basic-user",
    "description": "Basic user with read-only access",
    "permissions": ["users:read", "products:read"]
  }'
```

### Creating an API Key

```bash
curl -X POST http://localhost:3000/api/admin/api-keys \
  -H "X-Root-API-Key: orbitkey_root_your_root_key" \
  -H "Content-Type: application/json" \
  -d '{
    "role_id": 1,
    "description": "API key for Example App",
    "expires_in": 30
  }'
```

### Using an API Key

```bash
curl http://localhost:3000/api/users \
  -H "X-API-Key: orbitkey_your_api_key"
```

## License

MIT 