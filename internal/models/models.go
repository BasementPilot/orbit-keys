// Package models defines the data structures used throughout the OrbitKeys API key management system.
// It provides the main entities (Role and APIKey) along with their relationships and business logic
// for permission checking and validation.
package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Role represents a role with a set of specific permissions that can be assigned to API keys.
// Each role has a unique name and a list of permissions that determine what actions
// API keys with this role can perform.
type Role struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"unique;not null"`
	Description string         `json:"description"`
	Permissions string         `json:"permissions" gorm:"type:text"` // Stored as comma-separated string
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	APIKeys     []APIKey       `json:"-" gorm:"foreignKey:RoleID"`
}

// APIKey represents an API key used for authentication and authorization in the system.
// Each API key is associated with a role that determines its permissions.
// API keys can have an optional expiration date and track when they were last used.
type APIKey struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Key         string         `json:"key" gorm:"unique;not null;index"`
	RoleID      uint           `json:"role_id" gorm:"not null"`
	Role        Role           `json:"role" gorm:"constraint:OnDelete:CASCADE;"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	LastUsedAt  *time.Time     `json:"last_used_at"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// GetPermissions returns a slice of permissions for the role by parsing
// the comma-separated permissions string.
// Returns an empty slice if no permissions are assigned.
func (r *Role) GetPermissions() []string {
	return ParsePermissions(r.Permissions)
}

// HasPermission checks if the role has the specified permission.
// It supports wildcard permissions and resource-specific wildcards.
// Returns true if the role has the permission, false otherwise.
func (r *Role) HasPermission(permission string) bool {
	permissions := r.GetPermissions()
	return CheckPermission(permission, permissions)
}

// AddPermission adds a new permission to the role if it's not already present.
// The permission is validated to ensure it follows the correct format before being added.
// If the permission is invalid or already exists, no changes are made.
func (r *Role) AddPermission(permission string) {
	if !ValidatePermissionFormat(permission) {
		return
	}
	
	currentPerms := r.GetPermissions()
	if contains(currentPerms, permission) {
		return // Already has this permission
	}
	
	if r.Permissions == "" {
		r.Permissions = permission
	} else {
		r.Permissions = r.Permissions + "," + permission
	}
}

// RemovePermission removes a permission from the role.
// If the permission doesn't exist in the role, no changes are made.
func (r *Role) RemovePermission(permission string) {
	currentPerms := r.GetPermissions()
	newPerms := make([]string, 0)
	
	for _, p := range currentPerms {
		if p != permission {
			newPerms = append(newPerms, p)
		}
	}
	
	r.Permissions = strings.Join(newPerms, ",")
}

// IsExpired checks if the API key has expired by comparing its expiration date with the current time.
// Returns true if the key has an expiration date set and that date is in the past.
// Returns false if the key has no expiration date or if the expiration date is in the future.
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return k.ExpiresAt.Before(time.Now())
}

// UpdateLastUsed updates the LastUsedAt field of the API key to the current time.
// This is called whenever an API key is used for authentication to track usage.
// Returns an error if the database update fails.
func (k *APIKey) UpdateLastUsed(db *gorm.DB) error {
	now := time.Now()
	k.LastUsedAt = &now
	return db.Model(k).Update("last_used_at", now).Error
} 