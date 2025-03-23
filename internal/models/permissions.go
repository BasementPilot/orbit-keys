package models

import (
	"strings"
)

// Permission constants
const (
	WildcardPermission = "*"
	PermissionSeparator = ":"
)

// ParsePermissions converts a comma-separated permissions string to a slice
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

// CheckPermission checks if a required permission is included in the permissions slice
// It supports wildcard permissions (*, resource:*)
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

// FormatPermission formats a permission string as "resource:action"
func FormatPermission(resource, action string) string {
	return resource + PermissionSeparator + action
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidatePermissionFormat validates that a permission is correctly formatted
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