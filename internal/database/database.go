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

// DB is the global database instance
var DB *gorm.DB

// InitDB initializes the database connection
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

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// CreateDefaultAdminRole creates a default admin role if it doesn't exist
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

// CloseDB closes the database connection
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