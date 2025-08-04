package main

import (
	"flag"
	"fmt"
	"os"

	"linke/internal/application/bootstrap"
	"linke/internal/shared/database"

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

	// 如果是迁移命令，直接处理后退出
	if *runMigration || *migrateCommand != "" {
		migrationCLI := database.NewMigrationCLI()
		migrationCLI.HandleMigrationCommand(*runMigration, *migrateCommand, *migrateVersion, *migrateSteps)
		return
	}

	// 创建并启动应用
	app := bootstrap.NewApplication()

	// 运行应用
	if err := app.Err(); err != nil {
		fmt.Printf("❌ Application startup failed\n")
		fmt.Printf("   Error: %v\n", err)
		fmt.Printf("   This error occurred during dependency injection initialization.\n")
		fmt.Printf("   Common causes:\n")
		fmt.Printf("   - Database connection failure (check DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)\n")
		fmt.Printf("   - Redis connection failure (check REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_DB)\n")
		fmt.Printf("   - Invalid JWT_SECRET (must be 32+ characters)\n")
		fmt.Printf("   - Missing required environment variables\n")
		fmt.Printf("   - Configuration validation errors\n")
		fmt.Printf("\n   For detailed troubleshooting, run: make security-check\n")
		os.Exit(1)
	}
	
	app.Run()
}