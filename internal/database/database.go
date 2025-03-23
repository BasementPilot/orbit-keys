// Package database handles database connections and operations for the OrbitKeys system.
// It provides functions for initializing the SQLite database, running migrations,
// and creating default data needed for the system to function properly.
package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BasementPilot/orbit-keys/config"
	"github.com/BasementPilot/orbit-keys/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance used throughout the application.
// It should be initialized with InitDB before use.
var DB *gorm.DB

// InitDB initializes the database connection and performs initial setup.
// It creates the database file if it doesn't exist, sets up the connection
// with appropriate logging, and runs necessary migrations.
//
// The cfg parameter provides database configuration, particularly the DBPath.
// Returns an error if any part of the initialization process fails.
func InitDB(cfg *config.Config) error {
	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = "orbitkeys.db" // Default path
	}

	// Create the database directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Set up logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Open database connection
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	err = DB.AutoMigrate(&models.Role{}, &models.APIKey{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// GetDB returns the global database instance.
// The database must be initialized with InitDB before calling this function.
func GetDB() *gorm.DB {
	return DB
}

// CreateDefaultAdminRole ensures that an admin role with full permissions exists in the database.
// If the admin role already exists, this function does nothing.
// If it doesn't exist, it creates a new role with the name "admin" and wildcard (*) permission.
//
// Returns an error if database operations fail.
func CreateDefaultAdminRole() error {
	var count int64
	if err := DB.Model(&models.Role{}).Where("name = ?", "admin").Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		adminRole := models.Role{
			Name:        "admin",
			Description: "Administrator role with full access",
			Permissions: "*", // Wildcard permission for full access
		}

		if err := DB.Create(&adminRole).Error; err != nil {
			return fmt.Errorf("failed to create admin role: %w", err)
		}
		log.Println("Default admin role created")
	}

	return nil
}

// CloseDB properly closes the database connection when the application exits.
// This should be called as part of shutdown procedures to ensure all database
// operations are completed and resources are released.
func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Println("Error getting underlying SQL DB:", err)
			return
		}
		err = sqlDB.Close()
		if err != nil {
			log.Println("Error closing database connection:", err)
		}
	}
} 