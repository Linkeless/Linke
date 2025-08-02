package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	OAuth2   OAuth2Config
	JWT      JWTConfig
	Log      LogConfig
	API      APIConfig
}

type ServerConfig struct {
	Port string
}

type APIConfig struct {
	ServerToken string // Token for server API authentication
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type OAuth2Config struct {
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURL   string
	GitHubClientID      string
	GitHubClientSecret  string
	GitHubRedirectURL   string
	TelegramBotToken    string
	TelegramRedirectURL string
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

type LogConfig struct {
	Level  string
	Format string
	Output string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		// Use standard log here since logger might not be initialized yet
		// This is informational only, not an error
		fmt.Println("INFO: No .env file found, using environment variables")
	}

	// Critical: JWT_SECRET must be provided and secure
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		// Keep log.Fatal for critical security issues that should prevent startup
		log.Fatal("SECURITY ERROR: JWT_SECRET environment variable is required and cannot be empty. " +
			"This is a critical security requirement. Generate a secure random key: " +
			"openssl rand -hex 32")
	}

	// Validate JWT secret strength
	if len(jwtSecret) < 32 {
		log.Fatal("SECURITY ERROR: JWT_SECRET must be at least 32 characters long for security. " +
			"Current length: " + strconv.Itoa(len(jwtSecret)) + ". " +
			"Generate a secure random key: openssl rand -hex 32")
	}

	// Additional validation: ensure it's not the old default weak key
	if jwtSecret == "your-super-secret-jwt-key" || jwtSecret == "your-super-secret-jwt-key-make-it-strong-and-long" {
		log.Fatal("SECURITY ERROR: Cannot use the default JWT_SECRET value. " +
			"This is a known weak key that compromises system security. " +
			"Generate a secure random key: openssl rand -hex 32")
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "linke"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		OAuth2: OAuth2Config{
			GoogleClientID:      getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret:  getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirectURL:   getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
			GitHubClientID:      getEnv("GITHUB_CLIENT_ID", ""),
			GitHubClientSecret:  getEnv("GITHUB_CLIENT_SECRET", ""),
			GitHubRedirectURL:   getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/auth/github/callback"),
			TelegramBotToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
			TelegramRedirectURL: getEnv("TELEGRAM_REDIRECT_URL", "http://localhost:8080/auth/telegram/callback"),
		},
		JWT: JWTConfig{
			Secret:      jwtSecret,
			ExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "text"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		API: APIConfig{
			ServerToken: getEnv("SERVER_API_TOKEN", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GenerateSecureJWTKey generates a cryptographically secure JWT key
// This is a utility function for deployment and testing purposes
func GenerateSecureJWTKey() string {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		// Keep log.Fatal for cryptographic failures that indicate system problems
		log.Fatal("CRITICAL: Failed to generate secure random bytes: " + err.Error())
	}
	return hex.EncodeToString(bytes)
}
