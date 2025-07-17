package middleware

import (
	"strings"

	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	AuthContextKey = "auth_user"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header",
				logger.String("path", c.Request.URL.Path),
			)
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			logger.Warn("Invalid authorization header format",
				logger.String("path", c.Request.URL.Path),
			)
			response.Unauthorized(c, "Invalid authorization header format. Use 'Bearer <token>'")
			c.Abort()
			return
		}

		token := tokenParts[1]
		user, err := authService.ValidateToken(token)
		if err != nil {
			logger.Warn("Invalid token",
				logger.String("path", c.Request.URL.Path),
				logger.Error2("error", err),
			)
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Store user in context for use in handlers
		c.Set(AuthContextKey, user)
		c.Next()
	}
}

// OptionalAuthMiddleware creates a middleware that sets user context if token is present but doesn't require it
func OptionalAuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		token := tokenParts[1]
		user, err := authService.ValidateToken(token)
		if err != nil {
			// Don't fail the request, just continue without user context
			c.Next()
			return
		}

		// Store user in context for use in handlers
		c.Set(AuthContextKey, user)
		c.Next()
	}
}