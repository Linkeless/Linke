package database

import (
	"context"
	"fmt"
	"time"

	"linke/internal/shared/config"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Database wraps database connections
type Database struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// ConnectionConfig contains database connection configuration
type ConnectionConfig struct {
	MySQL MySQLConfig
	Redis RedisConfig
}

// MySQLConfig contains MySQL connection settings
type MySQLConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewDatabase creates a new Database instance with connections
func NewDatabase(cfg *config.Config) (*Database, error) {
	db, err := initMySQL(MySQLConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.Name,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MySQL: %w", err)
	}

	redisClient := initRedis(RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return &Database{
		DB:    db,
		Redis: redisClient,
	}, nil
}

// NewDatabaseWithConfig creates a new Database instance with custom configuration
func NewDatabaseWithConfig(connCfg ConnectionConfig) (*Database, error) {
	db, err := initMySQL(connCfg.MySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MySQL: %w", err)
	}

	redisClient := initRedis(connCfg.Redis)

	return &Database{
		DB:    db,
		Redis: redisClient,
	}, nil
}

// initMySQL initializes MySQL connection
func initMySQL(cfg MySQLConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

// initRedis initializes Redis connection
func initRedis(cfg RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return client
}

// TestConnections tests both MySQL and Redis connections
func (d *Database) TestConnections(ctx context.Context) error {
	// Test MySQL connection
	if err := d.TestMySQLConnection(ctx); err != nil {
		return fmt.Errorf("MySQL connection test failed: %w", err)
	}

	// Test Redis connection
	if err := d.TestRedisConnection(ctx); err != nil {
		return fmt.Errorf("Redis connection test failed: %w", err)
	}

	return nil
}

// TestMySQLConnection tests MySQL connection
func (d *Database) TestMySQLConnection(ctx context.Context) error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}

// TestRedisConnection tests Redis connection
func (d *Database) TestRedisConnection(ctx context.Context) error {
	_, err := d.Redis.Ping(ctx).Result()
	return err
}

// Close closes all database connections
func (d *Database) Close() error {
	var errs []error

	// Close Redis connection
	if d.Redis != nil {
		if err := d.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis connection: %w", err))
		}
	}

	// Close MySQL connection
	if d.DB != nil {
		sqlDB, err := d.DB.DB()
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get SQL DB: %w", err))
		} else {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close MySQL connection: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors occurred while closing connections: %v", errs)
	}

	return nil
}

// GetDB returns the GORM database instance
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// GetRedis returns the Redis client instance
func (d *Database) GetRedis() *redis.Client {
	return d.Redis
}

// WithTransaction executes a function within a database transaction
func (d *Database) WithTransaction(fn func(*gorm.DB) error) error {
	return d.DB.Transaction(fn)
}

// HealthCheck performs a health check on both connections
func (d *Database) HealthCheck(ctx context.Context) map[string]bool {
	result := make(map[string]bool)

	// Check MySQL
	result["mysql"] = d.TestMySQLConnection(ctx) == nil

	// Check Redis
	result["redis"] = d.TestRedisConnection(ctx) == nil

	return result
}