package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Role represents a role with specific permissions
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

// APIKey represents an API key used for authentication
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

// GetPermissions returns a slice of permissions for the role
func (r *Role) GetPermissions() []string {
	return ParsePermissions(r.Permissions)
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(permission string) bool {
	permissions := r.GetPermissions()
	return CheckPermission(permission, permissions)
}

// AddPermission adds a new permission to the role
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

// RemovePermission removes a permission from the role
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

// IsExpired checks if the API key has expired
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return k.ExpiresAt.Before(time.Now())
}

// UpdateLastUsed updates the LastUsedAt field to the current time
func (k *APIKey) UpdateLastUsed(db *gorm.DB) error {
	now := time.Now()
	k.LastUsedAt = &now
	return db.Model(k).Update("last_used_at", now).Error
} 