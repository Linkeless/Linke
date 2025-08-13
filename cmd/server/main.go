package main

import (
	"flag"
	"fmt"
	"os"

	"linke/internal/application/bootstrap"
	"linke/internal/shared/database"
	"linke/internal/shared/logger"

	// Swagger 文档
	_ "linke/docs"
)

// @title Linke API
// @version 1.0
// @description A comprehensive service management platform with subscription-based billing, user management, and server administration. Features include OAuth2 authentication, traffic subscription management, multi-gateway payments, referral programs, and customer support system.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Initialize logger early
	logConfig := logger.LogConfig{
		Level:  logger.GetEnvLogLevel(),
		Format: logger.GetEnvLogFormat(),
		Output: logger.GetEnvLogOutput(),
	}

	if err := logger.InitLogger(logConfig); err != nil {
		// Use fmt.Printf for pre-logger errors since logger isn't initialized yet
		fmt.Printf("❌ Failed to initialize logger: %v\n", err)
		fmt.Printf("   Falling back to default logger configuration\n")
		os.Exit(1)
	}

	// Parse command line flags
	var (
		runMigration    = flag.Bool("migrate", false, "Run database migrations and exit (alias for -migrate-command=up)")
		migrateCommand  = flag.String("migrate-command", "", "Migration command (up, down, reset, status, force, goto, steps, list, fix-dirty)")
		migrateVersion  = flag.String("migrate-version", "", "Target version for goto/force commands")
		migrateSteps    = flag.String("migrate-steps", "", "Number of steps for steps command")
		showMigrateHelp = flag.Bool("migrate-help", false, "Show migration help and exit")
	)
	flag.Parse()

	// Show migration help if requested
	if *showMigrateHelp {
		migrationCLI := database.NewMigrationCLI()
		migrationCLI.ShowMigrationHelp()
		return
	}

	// Handle migration commands
	if *runMigration || *migrateCommand != "" {
		migrationCLI := database.NewMigrationCLI()
		migrationCLI.HandleMigrationCommand(*runMigration, *migrateCommand, *migrateVersion, *migrateSteps)
		return
	}

	// Create and start application
	app := bootstrap.NewApplication()

	// Check for application initialization errors
	if err := app.Err(); err != nil {
		logger.Fatal("Application startup failed",
			logger.ErrorField(err),
			logger.String("troubleshooting_hint", "Common causes: database connection failure, Redis connection failure, invalid JWT_SECRET, missing environment variables, configuration validation errors"),
			logger.String("troubleshooting_command", "make security-check"),
		)
		os.Exit(1)
	}

	// Run the application
	app.Run()
}
