package models

import (
	"testing"
	"time"
)

func TestRolePermissions(t *testing.T) {
	// Create a role with specific permissions
	role := Role{
		Name:        "test-role",
		Description: "Test role",
		Permissions: "users:read,products:write,admin:*",
	}
	
	// Test GetPermissions
	perms := role.GetPermissions()
	if len(perms) != 3 {
		t.Errorf("Expected 3 permissions but got %d", len(perms))
	}
	
	// Manually check permission validation
	// Note: This is dependent on the functions in permissions.go
	validPermissions := []string{"users:read", "products:write", "admin:*"}
	for _, p := range validPermissions {
		if !ValidatePermissionFormat(p) {
			t.Errorf("Permission %s should be valid", p)
		}
	}
	
	invalidPermissions := []string{"users", ":", "users:", ":read", ""}
	for _, p := range invalidPermissions {
		if ValidatePermissionFormat(p) {
			t.Errorf("Permission %s should be invalid", p)
		}
	}
}

func TestAPIKeyExpiration(t *testing.T) {
	// Create a key that's not expired
	futureTime := time.Now().Add(24 * time.Hour)
	nonExpiredKey := APIKey{
		Key:         "test_key",
		ExpiresAt:   &futureTime,
	}
	
	if nonExpiredKey.IsExpired() {
		t.Error("Key should not be expired")
	}
	
	// Create a key that's expired
	pastTime := time.Now().Add(-24 * time.Hour)
	expiredKey := APIKey{
		Key:         "test_key_expired",
		ExpiresAt:   &pastTime,
	}
	
	if !expiredKey.IsExpired() {
		t.Error("Key should be expired")
	}
	
	// Create a key with no expiration
	noExpirationKey := APIKey{
		Key:         "test_key_no_expiration",
		ExpiresAt:   nil,
	}
	
	if noExpirationKey.IsExpired() {
		t.Error("Key with no expiration should not be expired")
	}
} 