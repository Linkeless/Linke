package migration

import (
	"os"
	"strings"

	"linke/internal/logger"
	"linke/internal/model"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	// Check if migration should run
	runMigration := os.Getenv("RUN_MIGRATION")
	if strings.ToLower(runMigration) != "true" {
		logger.Info("Database migration skipped (set RUN_MIGRATION=true to enable)")
		return nil
	}

	logger.Info("Starting database migration")

	// Migrate User model
	if err := db.AutoMigrate(&model.User{}); err != nil {
		logger.Error("Failed to migrate User model", logger.Error2("error", err))
		return err
	}

	// Migrate InviteCode model
	if err := db.AutoMigrate(&model.InviteCode{}); err != nil {
		logger.Error("Failed to migrate InviteCode model", logger.Error2("error", err))
		return err
	}

	// Migrate InviteCodeUsage model
	if err := db.AutoMigrate(&model.InviteCodeUsage{}); err != nil {
		logger.Error("Failed to migrate InviteCodeUsage model", logger.Error2("error", err))
		return err
	}

	logger.Info("Database migration completed successfully")
	return nil
}