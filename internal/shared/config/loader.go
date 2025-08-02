package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigLoader provides configuration loading functionality
type ConfigLoader struct {
	envFile string
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(envFile string) *ConfigLoader {
	return &ConfigLoader{
		envFile: envFile,
	}
}

// LoadFromFile loads configuration from a specific file
func (cl *ConfigLoader) LoadFromFile(path string) (*Config, error) {
	// Set environment file path if specified
	if path != "" {
		os.Setenv("ENV_FILE", path)
	}
	
	return LoadConfig(), nil
}

// LoadFromEnvironment loads configuration from environment variables only
func (cl *ConfigLoader) LoadFromEnvironment() (*Config, error) {
	// Clear any .env file setting to force environment variables only
	os.Unsetenv("ENV_FILE")
	return LoadConfig(), nil
}

// ValidateConfig validates the loaded configuration
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Validate database configuration
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	// Validate Redis configuration
	if cfg.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	// Validate JWT configuration
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	return nil
}

// GetConfigDir returns the directory containing configuration files
func GetConfigDir() (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Look for config directory in current directory or parent directories
	configDir := filepath.Join(cwd, "config")
	if _, err := os.Stat(configDir); err == nil {
		return configDir, nil
	}

	// Check parent directory
	parentDir := filepath.Dir(cwd)
	configDir = filepath.Join(parentDir, "config")
	if _, err := os.Stat(configDir); err == nil {
		return configDir, nil
	}

	return "", fmt.Errorf("config directory not found")
}

// GetEnvFile returns the path to the .env file
func GetEnvFile() string {
	envFile := os.Getenv("ENV_FILE")
	if envFile != "" {
		return envFile
	}

	// Default to .env in current directory
	cwd, err := os.Getwd()
	if err != nil {
		return ".env"
	}

	return filepath.Join(cwd, ".env")
}

// SetConfigForTesting sets configuration values for testing purposes
func SetConfigForTesting() *Config {
	os.Setenv("JWT_SECRET", GenerateSecureJWTKey())
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_NAME", "linke_test")
	os.Setenv("REDIS_HOST", "localhost")
	
	return LoadConfig()
}

// ClearTestingConfig clears testing configuration
func ClearTestingConfig() {
	testVars := []string{
		"JWT_SECRET",
		"DB_HOST",
		"DB_NAME", 
		"REDIS_HOST",
	}
	
	for _, variable := range testVars {
		os.Unsetenv(variable)
	}
}