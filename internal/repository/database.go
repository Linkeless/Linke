package repository

import (
	"context"
	"fmt"
	"time"

	"linke/config"
	"linke/internal/logger"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	db, err := initMySQL(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MySQL: %w", err)
	}

	redisClient := initRedis(cfg.Redis)

	return &Database{
		DB:    db,
		Redis: redisClient,
	}, nil
}

func initMySQL(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func initRedis(cfg config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		logger.Error("Failed to connect to Redis", logger.Error2("error", err))
	} else {
		logger.Info("Connected to Redis successfully",
			logger.String("host", cfg.Host),
			logger.String("port", cfg.Port),
		)
	}

	return client
}

func (d *Database) Close() error {
	if d.Redis != nil {
		if err := d.Redis.Close(); err != nil {
			logger.Error("Failed to close Redis connection", logger.Error2("error", err))
		} else {
			logger.Info("Redis connection closed")
		}
	}

	if d.DB != nil {
		sqlDB, err := d.DB.DB()
		if err != nil {
			return err
		}
		if err := sqlDB.Close(); err != nil {
			logger.Error("Failed to close database connection", logger.Error2("error", err))
			return err
		}
		logger.Info("Database connection closed")
	}

	return nil
}