package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"linke/config"
	"linke/internal/logger"
	"linke/internal/migration"
	"linke/internal/modules"
	"linke/internal/queue"
	"linke/internal/repository"
	"linke/internal/routes"

	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	
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
		showMigrationHelp()
		return
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	if err := logger.InitLogger(logger.LogConfig{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	}); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("Starting Linke server",
		logger.String("version", "1.0"),
		logger.String("log_level", cfg.Log.Level),
		logger.String("log_format", cfg.Log.Format),
	)

	// Initialize database
	db, err := repository.NewDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", logger.Error2("error", err))
	}
	defer db.Close()

	// Handle migration commands
	if *runMigration || *migrateCommand != "" {
		migrationService := migration.NewMigrationService(
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
		)

		// Validate database connection first
		if err := migrationService.ValidateConnection(); err != nil {
			logger.Fatal("Database connection failed", logger.Error2("error", err))
		}

		// Determine command to execute
		command := *migrateCommand
		if *runMigration && command == "" {
			command = "up"
		}

		if err := executeMigrationCommand(migrationService, command, *migrateVersion, *migrateSteps); err != nil {
			logger.Fatal("Migration command failed", logger.Error2("error", err), logger.String("command", command))
		}

		logger.Info("Migration command completed, exiting...", logger.String("command", command))
		return
	}

	// Initialize task queue
	taskQueue := queue.NewTaskQueue(db.Redis)
	processor := queue.NewTaskProcessor(db.Redis)
	processor.RegisterHandler("email", queue.EmailTaskHandler)
	processor.RegisterHandler("notification", queue.NotificationTaskHandler)
	processor.RegisterHandler("data_processing", queue.DataProcessingTaskHandler)
	
	// TODO: Add node data processing handlers when needed

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := processor.Start(ctx); err != nil {
			logger.ErrorWithRateLimit("task_processor_error", "Task processor failed", logger.Error2("error", err))
		}
	}()

	// Initialize modules using the simple manager
	moduleManager := modules.NewSimpleManager(cfg, db, taskQueue)
	
	// Initialize route manager
	routeManager := routes.NewSimpleManager(moduleManager)
	routeManager.SetupRoutes()

	// Get router and add swagger
	router := routeManager.GetRouter()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	// Start server
	go func() {
		logger.Info("Server starting", logger.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", logger.Error2("error", err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Stop task processor
	cancel()
	processor.Stop()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", logger.Error2("error", err))
	}

	// Close task queue
	if err := taskQueue.Close(); err != nil {
		logger.Error("Failed to close task queue", logger.Error2("error", err))
	}

	logger.Info("Server exited")
}

// Migration helper functions (same as original)
func showMigrationHelp() {
	fmt.Println("Database Migration Commands")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/server/main.go [migration-options]")
	fmt.Println("")
	fmt.Println("Migration Options:")
	fmt.Println("  -migrate                    Run database migrations and exit (same as -migrate-command=up)")
	fmt.Println("  -migrate-command=<command>  Execute specific migration command")
	fmt.Println("  -migrate-version=<N>        Target version for goto/force commands")
	fmt.Println("  -migrate-steps=<N>          Number of steps for steps command")
	fmt.Println("  -migrate-help               Show this help and exit")
	fmt.Println("")
	fmt.Println("Migration Commands:")
	fmt.Println("  up       - Run all pending migrations")
	fmt.Println("  down     - Rollback one migration")
	fmt.Println("  reset    - Drop all tables and re-run migrations (DANGEROUS!)")
	fmt.Println("  status   - Show current migration version")
	fmt.Println("  list     - List all applied migrations")
	fmt.Println("  force    - Force set migration version (use with caution)")
	fmt.Println("  goto     - Migrate to specific version")
	fmt.Println("  steps    - Run specific number of migration steps")
	fmt.Println("  fix-dirty - Fix dirty migration state (requires -migrate-version)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/server/main.go -migrate")
	fmt.Println("  go run cmd/server/main.go -migrate-command=status")
	fmt.Println("  go run cmd/server/main.go -migrate-command=fix-dirty -migrate-version=9")
	fmt.Println("  go run cmd/server/main.go -migrate-command=goto -migrate-version=5")
	fmt.Println("  go run cmd/server/main.go -migrate-command=steps -migrate-steps=2")
	fmt.Println("")
	fmt.Println("Fixing Dirty Migration State:")
	fmt.Println("  If you see 'Dirty database version X' error:")
	fmt.Println("  go run cmd/server/main.go -migrate-command=fix-dirty -migrate-version=X")
}

func executeMigrationCommand(migrationService *migration.MigrationService, command, version, steps string) error {
	switch command {
	case "up":
		logger.Info("Running database migrations...")
		return migrationService.Up()

	case "down":
		logger.Info("Rolling back one migration...")
		return migrationService.Down()

	case "reset":
		fmt.Print("WARNING: This will drop all tables and re-run migrations. Are you sure? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm == "y" || confirm == "Y" {
			logger.Info("Resetting database...")
			return migrationService.Reset()
		} else {
			logger.Info("Reset cancelled")
			return nil
		}

	case "status":
		status, err := migrationService.Status()
		if err != nil {
			return err
		}
		fmt.Println(status)
		return nil

	case "list":
		versions, err := migrationService.GetAppliedMigrations()
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			fmt.Println("No migrations have been applied")
		} else {
			fmt.Println("Applied migrations:")
			for _, v := range versions {
				fmt.Printf("  - Version %d\n", v)
			}
		}
		return nil

	case "force":
		if version == "" {
			return fmt.Errorf("version is required for force command")
		}
		v, err := strconv.Atoi(version)
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}
		logger.Warn("Forcing migration version", logger.Int("version", v))
		return migrationService.Force(v)

	case "goto":
		if version == "" {
			return fmt.Errorf("version is required for goto command")
		}
		v, err := strconv.ParseUint(version, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}
		logger.Info("Migrating to specific version", logger.Uint("version", uint(v)))
		return migrationService.Goto(uint(v))

	case "steps":
		if steps == "" {
			return fmt.Errorf("steps is required for steps command")
		}
		s, err := strconv.Atoi(steps)
		if err != nil {
			return fmt.Errorf("invalid steps number: %w", err)
		}
		direction := "up"
		if s < 0 {
			direction = "down"
		}
		logger.Info("Running migration steps", logger.Int("steps", s), logger.String("direction", direction))
		return migrationService.Steps(s)

	case "fix-dirty":
		if version == "" {
			return fmt.Errorf("version is required for fix-dirty command. Use the version shown in the dirty database error")
		}
		v, err := strconv.Atoi(version)
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}
		
		fmt.Printf("WARNING: This will force the migration version to %d without running the migration.\n", v)
		fmt.Print("This should only be used to fix a dirty migration state. Continue? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm == "y" || confirm == "Y" {
			logger.Warn("Fixing dirty migration state", logger.Int("version", v))
			if err := migrationService.Force(v); err != nil {
				return fmt.Errorf("failed to fix dirty migration: %w", err)
			}
			logger.Info("Dirty migration state fixed", logger.Int("version", v))
			fmt.Println("Migration state fixed. You can now run migrations again.")
			return nil
		} else {
			logger.Info("Fix dirty operation cancelled")
			return nil
		}

	case "":
		return fmt.Errorf("migration command is required")

	default:
		return fmt.Errorf("unknown migration command: %s", command)
	}
}