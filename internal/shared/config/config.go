package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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
	Payment  PaymentSecurityConfig
	Cache    CacheConfig
	Versioning VersioningConfig
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

type PaymentSecurityConfig struct {
	// Signature verification
	RequireSignature          bool     `json:"require_signature"`
	EpaySignKey              string   `json:"epay_sign_key"`
	EpusdtSignKey            string   `json:"epusdt_sign_key"`
	
	// IP whitelist
	EnableIPWhitelist        bool     `json:"enable_ip_whitelist"`
	EpayIPWhitelist          []string `json:"epay_ip_whitelist"`
	EpusdtIPWhitelist        []string `json:"epusdt_ip_whitelist"`
	
	// Replay attack prevention
	EnableReplayProtection   bool     `json:"enable_replay_protection"`
	ReplayTimeWindowMinutes  int      `json:"replay_time_window_minutes"`
	
	// Rate limiting
	NotifyRateLimit          int      `json:"notify_rate_limit"`          // requests per minute
	NotifyRateBurst          int      `json:"notify_rate_burst"`          // burst size
	
	// Security headers
	RequireHTTPS             bool     `json:"require_https"`
	MaxRequestSize           int64    `json:"max_request_size"`           // in bytes
}

type CacheConfig struct {
	DefaultTTL      int  `json:"default_ttl"`        // in seconds
	EnableMetrics   bool `json:"enable_metrics"`
	EnableDebugLog  bool `json:"enable_debug_log"`
}

type VersioningConfig struct {
	// Strategy defines how version is extracted from requests
	Strategy string `json:"strategy"` // url_path, header, query, content_type
	
	// DefaultVersion is used when no version is specified
	DefaultVersion string `json:"default_version"`
	
	// MinVersion is the minimum supported version
	MinVersion string `json:"min_version"`
	
	// MaxVersion is the maximum supported version  
	MaxVersion string `json:"max_version"`
	
	// HeaderName is the header name for header strategy
	HeaderName string `json:"header_name"`
	
	// QueryParam is the query parameter name for query strategy
	QueryParam string `json:"query_param"`
	
	// URLPrefix is the URL prefix for path strategy
	URLPrefix string `json:"url_prefix"`
	
	// EnableDeprecationHeaders adds deprecation headers to responses
	EnableDeprecationHeaders bool `json:"enable_deprecation_headers"`
	
	// EnableAutoMigration enables automatic version migration for compatible versions
	EnableAutoMigration bool `json:"enable_auto_migration"`
	
	// SunsetV1Date is the sunset date for version 1 (ISO 8601 format)
	SunsetV1Date string `json:"sunset_v1_date"`
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		// Use standard log here since logger might not be initialized yet
		// This is informational only, not an error
		fmt.Printf("INFO: No .env file found in current directory (%s), using environment variables only\n", 
			getCurrentDirectory())
	} else {
		fmt.Printf("INFO: Configuration loaded from .env file in %s\n", getCurrentDirectory())
	}

	// Critical: JWT_SECRET must be provided and secure
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		// Keep log.Fatal for critical security issues that should prevent startup
		log.Fatal("❌ SECURITY ERROR: JWT_SECRET environment variable is required and cannot be empty.\n" +
			"   This is a critical security requirement for token-based authentication.\n" +
			"   Solutions:\n" +
			"   1. Generate a secure random key: openssl rand -hex 32\n" +
			"   2. Add it to your .env file: JWT_SECRET=your_generated_key\n" +
			"   3. Or export it: export JWT_SECRET=your_generated_key\n" +
			"   Current working directory: " + getCurrentDirectory())
	}

	// Validate JWT secret strength
	if len(jwtSecret) < 32 {
		log.Fatal("❌ SECURITY ERROR: JWT_SECRET must be at least 32 characters long for security.\n" +
			fmt.Sprintf("   Current length: %d characters (minimum required: 32)\n", len(jwtSecret)) +
			"   Your current JWT_SECRET is too weak and compromises system security.\n" +
			"   Solutions:\n" +
			"   1. Generate a secure random key: openssl rand -hex 32\n" +
			"   2. Update your .env file or environment variable\n" +
			"   Current working directory: " + getCurrentDirectory())
	}

	// Additional validation: ensure it's not the old default weak key
	if jwtSecret == "your-super-secret-jwt-key" || jwtSecret == "your-super-secret-jwt-key-make-it-strong-and-long" {
		log.Fatal("❌ SECURITY ERROR: Cannot use the default JWT_SECRET value.\n" +
			"   This is a known weak key that compromises system security.\n" +
			"   The current JWT_SECRET appears to be a default/example value.\n" +
			"   Solutions:\n" +
			"   1. Generate a secure random key: openssl rand -hex 32\n" +
			"   2. Update your .env file with the new key\n" +
			"   3. Never use example/default keys in production\n" +
			"   Current working directory: " + getCurrentDirectory())
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
		Payment: PaymentSecurityConfig{
			RequireSignature:        getEnvBool("PAYMENT_REQUIRE_SIGNATURE", true),
			EpaySignKey:            getEnv("EPAY_SIGN_KEY", ""),
			EpusdtSignKey:          getEnv("EPUSDT_SIGN_KEY", ""),
			EnableIPWhitelist:      getEnvBool("PAYMENT_ENABLE_IP_WHITELIST", false),
			EpayIPWhitelist:        getEnvStringSlice("EPAY_IP_WHITELIST", []string{}),
			EpusdtIPWhitelist:      getEnvStringSlice("EPUSDT_IP_WHITELIST", []string{}),
			EnableReplayProtection: getEnvBool("PAYMENT_ENABLE_REPLAY_PROTECTION", true),
			ReplayTimeWindowMinutes: getEnvInt("PAYMENT_REPLAY_TIME_WINDOW_MINUTES", 5),
			NotifyRateLimit:        getEnvInt("PAYMENT_NOTIFY_RATE_LIMIT", 10),
			NotifyRateBurst:        getEnvInt("PAYMENT_NOTIFY_RATE_BURST", 2),
			RequireHTTPS:           getEnvBool("PAYMENT_REQUIRE_HTTPS", false),
			MaxRequestSize:         int64(getEnvInt("PAYMENT_MAX_REQUEST_SIZE", 1048576)), // 1MB default
		},
		Cache: CacheConfig{
			DefaultTTL:      getEnvInt("CACHE_DEFAULT_TTL", 300), // 5 minutes default
			EnableMetrics:   getEnvBool("CACHE_ENABLE_METRICS", false),
			EnableDebugLog:  getEnvBool("CACHE_ENABLE_DEBUG_LOG", false),
		},
		Versioning: VersioningConfig{
			Strategy:                 getEnv("API_VERSION_STRATEGY", "url_path"),
			DefaultVersion:           getEnv("API_DEFAULT_VERSION", "1.0.0"),
			MinVersion:               getEnv("API_MIN_VERSION", "1.0.0"),
			MaxVersion:               getEnv("API_MAX_VERSION", "2.0.0"),
			HeaderName:               getEnv("API_VERSION_HEADER", "X-API-Version"),
			QueryParam:               getEnv("API_VERSION_QUERY_PARAM", "version"),
			URLPrefix:                getEnv("API_URL_PREFIX", "/api"),
			EnableDeprecationHeaders: getEnvBool("API_ENABLE_DEPRECATION_HEADERS", true),
			EnableAutoMigration:      getEnvBool("API_ENABLE_AUTO_MIGRATION", false),
			SunsetV1Date:             getEnv("API_SUNSET_V1_DATE", "2025-12-31T23:59:59Z"),
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

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Split by comma and trim spaces
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}

// getCurrentDirectory returns current working directory for error messages
func getCurrentDirectory() string {
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "unknown"
}

// GenerateSecureJWTKey generates a cryptographically secure JWT key
// This is a utility function for deployment and testing purposes
func GenerateSecureJWTKey() string {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		// Keep log.Fatal for cryptographic failures that indicate system problems
		log.Fatal("❌ CRITICAL: Failed to generate secure random bytes: " + err.Error() + 
			"\n   This indicates a serious system problem with the cryptographic subsystem." +
			"\n   Please check your system's random number generator.")
	}
	return hex.EncodeToString(bytes)
}
