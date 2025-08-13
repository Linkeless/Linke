package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"linke/internal/shared/config"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:8080", "https://your-frontend-domain.com"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-CSRF-Token",
			"Cache-Control",
			"Accept-Encoding",
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORS returns a CORS middleware with default configuration
func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware with custom configuration
func CORSWithConfig(config *CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowedOrigin := "*"
		if len(config.AllowOrigins) > 0 {
			allowedOrigin = ""
			for _, configOrigin := range config.AllowOrigins {
				if configOrigin == "*" || configOrigin == origin {
					allowedOrigin = origin
					break
				}
			}
			if allowedOrigin == "" {
				allowedOrigin = config.AllowOrigins[0] // fallback to first allowed origin
			}
		}

		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CORSFromConfig creates CORS middleware from application config
func CORSFromConfig(cfg *config.Config) gin.HandlerFunc {
	corsConfig := DefaultCORSConfig()

	// Allow localhost origins for development
	// Include localhost origins regardless of log level for local development
	corsConfig.AllowOrigins = []string{
		"http://localhost:3000",
		"http://localhost:8080",
		"http://localhost:3001",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:8080",
		"http://127.0.0.1:3001",
		"https://localhost:3000",
		"https://localhost:8080",
		"https://localhost:3001",
	}

	// In production deployment, add production domains
	if cfg.Log.Level != "debug" {
		// Add production domains while keeping localhost for development
		corsConfig.AllowOrigins = append(corsConfig.AllowOrigins, []string{
			"https://your-frontend-domain.com",
			"https://www.your-frontend-domain.com",
		}...)
	}

	return CORSWithConfig(corsConfig)
}
