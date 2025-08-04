package database

import (
	"fmt"
	"os"
	"strconv"

	"linke/internal/shared/config"
	loggerPkg "linke/internal/shared/logger"
)

// MigrationCLI handles database migration commands
type MigrationCLI struct {
	migrationService *MigrationService
}

// NewMigrationCLI creates a new migration CLI handler
func NewMigrationCLI() *MigrationCLI {
	return &MigrationCLI{}
}

// HandleMigrationCommand processes migration commands
func (cli *MigrationCLI) HandleMigrationCommand(runMigration bool, migrateCommand, migrateVersion, migrateSteps string) {
	// 加载配置
	fmt.Println("Loading configuration for migration...")
	cfg := config.LoadConfig()
	if err := config.ValidateConfig(cfg); err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		fmt.Printf("   Please check your environment variables and .env file\n")
		os.Exit(1)
	}

	// 初始化日志
	if err := loggerPkg.InitLogger(loggerPkg.LogConfig{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	}); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer loggerPkg.Sync()

	loggerPkg.Info("Running migration command")

	// 创建迁移服务
	cli.migrationService = NewMigrationService(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	// 验证数据库连接
	loggerPkg.Info("Validating database connection",
		loggerPkg.String("host", cfg.Database.Host),
		loggerPkg.String("port", cfg.Database.Port),
		loggerPkg.String("database", cfg.Database.Name))
	if err := cli.migrationService.ValidateConnection(); err != nil {
		loggerPkg.Fatal("Database connection failed",
			loggerPkg.ErrorField(err),
			loggerPkg.String("host", cfg.Database.Host),
			loggerPkg.String("port", cfg.Database.Port),
			loggerPkg.String("database", cfg.Database.Name))
		fmt.Printf("❌ Database connection failed\n")
		fmt.Printf("   Host: %s:%s\n", cfg.Database.Host, cfg.Database.Port)
		fmt.Printf("   Database: %s\n", cfg.Database.Name)
		fmt.Printf("   User: %s\n", cfg.Database.User)
		fmt.Printf("   Error: %v\n", err)
		fmt.Printf("   Common causes:\n")
		fmt.Printf("   - Database server not running\n")
		fmt.Printf("   - Incorrect credentials (DB_USER, DB_PASSWORD)\n")
		fmt.Printf("   - Network connectivity issues\n")
		fmt.Printf("   - Database does not exist\n")
		os.Exit(1)
	}

	// 确定要执行的命令
	command := migrateCommand
	if runMigration && command == "" {
		command = "up"
	}

	// 执行迁移命令
	if err := cli.executeMigrationCommand(command, migrateVersion, migrateSteps); err != nil {
		loggerPkg.Fatal("Migration command failed",
			loggerPkg.ErrorField(err),
			loggerPkg.String("command", command),
			loggerPkg.String("version", migrateVersion),
			loggerPkg.String("steps", migrateSteps))
		fmt.Printf("❌ Migration command '%s' failed: %v\n", command, err)
		fmt.Printf("   For help with migration commands, run: go run cmd/server/main.go -migrate-help\n")
		os.Exit(1)
	}

	loggerPkg.Info("Migration command completed, exiting...", loggerPkg.String("command", command))
}

// ShowMigrationHelp displays migration command help
func (cli *MigrationCLI) ShowMigrationHelp() {
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

func (cli *MigrationCLI) executeMigrationCommand(command, version, steps string) error {
	switch command {
	case "up":
		loggerPkg.Info("Running database migrations...")
		return cli.migrationService.Up()

	case "down":
		loggerPkg.Info("Rolling back one migration...")
		return cli.migrationService.Down()

	case "reset":
		fmt.Print("WARNING: This will drop all tables and re-run migrations. Are you sure? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm == "y" || confirm == "Y" {
			loggerPkg.Info("Resetting database...")
			return cli.migrationService.Reset()
		} else {
			loggerPkg.Info("Reset cancelled")
			return nil
		}

	case "status":
		status, err := cli.migrationService.Status()
		if err != nil {
			return err
		}
		fmt.Println(status)
		return nil

	case "list":
		versions, err := cli.migrationService.GetAppliedMigrations()
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
		loggerPkg.Warn("Forcing migration version", loggerPkg.Int("version", v))
		return cli.migrationService.Force(v)

	case "goto":
		if version == "" {
			return fmt.Errorf("version is required for goto command")
		}
		v, err := strconv.ParseUint(version, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}
		loggerPkg.Info("Migrating to specific version", loggerPkg.Uint("version", uint(v)))
		return cli.migrationService.Goto(uint(v))

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
		loggerPkg.Info("Running migration steps", loggerPkg.Int("steps", s), loggerPkg.String("direction", direction))
		return cli.migrationService.Steps(s)

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
			loggerPkg.Warn("Fixing dirty migration state", loggerPkg.Int("version", v))
			if err := cli.migrationService.Force(v); err != nil {
				return fmt.Errorf("failed to fix dirty migration: %w", err)
			}
			loggerPkg.Info("Dirty migration state fixed", loggerPkg.Int("version", v))
			fmt.Println("Migration state fixed. You can now run migrations again.")
			return nil
		} else {
			loggerPkg.Info("Fix dirty operation cancelled")
			return nil
		}

	case "":
		return fmt.Errorf("migration command is required")

	default:
		return fmt.Errorf("unknown migration command: %s", command)
	}
}