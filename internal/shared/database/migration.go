package database

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrationService handles database migrations
type MigrationService struct {
	dbURL          string
	mysqlDSN       string
	migrationsPath string
}

// MigrationConfig contains migration configuration
type MigrationConfig struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	MigrationsPath string
}

// NewMigrationService creates a new migration service
func NewMigrationService(dbHost, dbPort, dbUser, dbPassword, dbName string) *MigrationService {
	dbURL := fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)

	return &MigrationService{
		dbURL:          dbURL,
		mysqlDSN:       mysqlDSN,
		migrationsPath: "migrations",
	}
}

// NewMigrationServiceWithConfig creates a migration service with custom config
func NewMigrationServiceWithConfig(cfg MigrationConfig) *MigrationService {
	dbURL := fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	migrationsPath := cfg.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	return &MigrationService{
		dbURL:          dbURL,
		mysqlDSN:       mysqlDSN,
		migrationsPath: migrationsPath,
	}
}

// createMigrate creates a new migrate instance
func (ms *MigrationService) createMigrate() (*migrate.Migrate, error) {
	// Get absolute path to migrations directory
	absPath, err := filepath.Abs(ms.migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create file source
	sourceURL := fmt.Sprintf("file://%s", absPath)

	// Create migrate instance
	m, err := migrate.New(sourceURL, ms.dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, nil
}

// Up runs all up migrations
func (ms *MigrationService) Up() error {
	m, err := ms.createMigrate()
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

// Down runs one down migration
func (ms *MigrationService) Down() error {
	m, err := ms.createMigrate()
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Steps(-1)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration rollback failed: %w", err)
	}

	return nil
}

// Reset drops all tables and re-runs all migrations
func (ms *MigrationService) Reset() error {
	m, err := ms.createMigrate()
	if err != nil {
		return err
	}
	defer m.Close()

	// Drop all tables
	err = m.Drop()
	if err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	// Re-run all migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to re-run migrations: %w", err)
	}

	return nil
}

// Version returns the current migration version
func (ms *MigrationService) Version() (uint, bool, error) {
	m, err := ms.createMigrate()
	if err != nil {
		return 0, false, err
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}

// Status returns migration status information
func (ms *MigrationService) Status() (string, error) {
	version, dirty, err := ms.Version()
	if err != nil {
		return "", err
	}

	if err == migrate.ErrNilVersion {
		return "No migrations applied", nil
	}

	status := fmt.Sprintf("Current version: %d", version)
	if dirty {
		status += " (dirty - migration failed)"
	}

	return status, nil
}

// Force sets the migration version without running migrations (use with caution)
func (ms *MigrationService) Force(version int) error {
	m, err := ms.createMigrate()
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Force(version)
	if err != nil {
		return fmt.Errorf("failed to force version: %w", err)
	}

	return nil
}

// Goto migrates to a specific version
func (ms *MigrationService) Goto(version uint) error {
	m, err := ms.createMigrate()
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Migrate(version)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	return nil
}

// Steps runs n migration steps (positive for up, negative for down)
func (ms *MigrationService) Steps(n int) error {
	m, err := ms.createMigrate()
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Steps(n)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run %d steps: %w", n, err)
	}

	return nil
}

// ValidateConnection tests the database connection
func (ms *MigrationService) ValidateConnection() error {
	// Use MySQL DSN format for direct connection
	db, err := sql.Open("mysql", ms.mysqlDSN)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// InitializeMigrationTable ensures golang-migrate can operate properly
// Note: golang-migrate automatically creates and manages the schema_migrations table
func (ms *MigrationService) InitializeMigrationTable() error {
	// Just validate that we can create a migrate instance
	// golang-migrate will handle schema_migrations table creation automatically
	m, err := ms.createMigrate()
	if err != nil {
		return fmt.Errorf("failed to initialize migration system: %w", err)
	}
	defer m.Close()

	// Check current version to ensure the migration system is working
	// This will automatically create the schema_migrations table if needed
	_, _, err = m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to check migration version: %w", err)
	}

	return nil
}

// GetAppliedMigrations returns a list of applied migration versions
func (ms *MigrationService) GetAppliedMigrations() ([]uint, error) {
	db, err := sql.Open("mysql", ms.mysqlDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		// If table doesn't exist, return nil slice
		return nil, nil
	}
	defer rows.Close()

	var versions []uint
	for rows.Next() {
		var version uint
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// IsMigrationTableDirty checks if there are any dirty migrations
func (ms *MigrationService) IsMigrationTableDirty() (bool, error) {
	db, err := sql.Open("mysql", ms.mysqlDSN)
	if err != nil {
		return false, fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE dirty = 1").Scan(&count)
	if err != nil {
		// If table doesn't exist, it's not dirty
		return false, nil
	}

	return count > 0, nil
}
