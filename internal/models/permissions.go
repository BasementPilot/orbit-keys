package models

import (
	"strings"
)

// Permission constants define special values used in the permission system.
const (
	// WildcardPermission represents a permission that grants access to all resources and actions.
	WildcardPermission = "*"
	
	// PermissionSeparator separates the resource from the action in permission strings.
	PermissionSeparator = ":"
)

// ParsePermissions converts a comma-separated permissions string to a slice of individual permission strings.
// Each permission is trimmed of whitespace. An empty input string returns an empty slice.
// This function is used to convert the database-stored permission format to a usable slice.
func ParsePermissions(permissions string) []string {
	if permissions == "" {
		return []string{}
	}
	
	// Split by comma and trim spaces
	perms := strings.Split(permissions, ",")
	for i, p := range perms {
		perms[i] = strings.TrimSpace(p)
	}
	
	return perms
}

// CheckPermission determines if a required permission is included in the user's permission set.
// It supports several permission matching patterns:
// 1. Wildcard (*) grants access to everything
// 2. Exact match between required and user permission
// 3. Resource wildcard (resource:*) grants access to all actions on a specific resource
//
// Returns true if the permission is granted, false otherwise.
func CheckPermission(requiredPermission string, userPermissions []string) bool {
	// If user has wildcard permission, they have access to everything
	for _, p := range userPermissions {
		if p == WildcardPermission {
			return true
		}
	}
	
	// Check specific permission match
	if contains(userPermissions, requiredPermission) {
		return true
	}
	
	// Check if user has wildcard for this resource
	parts := strings.Split(requiredPermission, PermissionSeparator)
	if len(parts) == 2 {
		resourceWildcard := parts[0] + PermissionSeparator + WildcardPermission
		return contains(userPermissions, resourceWildcard)
	}
	
	return false
}

// FormatPermission combines a resource name and action into a properly formatted permission string.
// The resulting format is "resource:action".
func FormatPermission(resource, action string) string {
	return resource + PermissionSeparator + action
}

// contains checks if a string slice contains a specific string.
// It's used internally for permission matching.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidatePermissionFormat checks if a permission string follows the correct format.
// Valid formats are:
// - Wildcard (*) for all permissions
// - "resource:action" where both resource and action are non-empty
// - "resource:*" for all actions on a specific resource
//
// Returns true if the format is valid, false otherwise.
func ValidatePermissionFormat(permission string) bool {
	// Wildcard is valid
	if permission == WildcardPermission {
		return true
	}
	
	// Check format: resource:action
	parts := strings.Split(permission, PermissionSeparator)
	if len(parts) != 2 {
		return false
	}
	
	// Both resource and action must not be empty unless it's a wildcard
	return (parts[0] != "" && (parts[1] != "" || parts[1] == WildcardPermission))
} 