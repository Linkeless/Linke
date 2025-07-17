package middleware

import (
	"linke/internal/model"
	"linke/internal/response"

	"github.com/gin-gonic/gin"
)

// RequireAdmin is a middleware that checks if the authenticated user has admin privileges
// This middleware should be used after the authentication middleware
func RequireAdmin() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get the user from context (set by auth middleware)
		userValue, exists := c.Get(AuthContextKey)
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		user, ok := userValue.(*model.User)
		if !ok {
			response.Unauthorized(c, "Invalid user context")
			c.Abort()
			return
		}

		// Check if user is admin
		if !user.IsAdmin() {
			response.Forbidden(c, "Admin privileges required")
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireAdminOrOwner is a middleware that checks if the user is admin or owns the resource
// This is useful for endpoints where users can access their own data or admins can access any data
func RequireAdminOrOwner(getUserIDFromPath func(*gin.Context) uint) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get the user from context (set by auth middleware)
		userValue, exists := c.Get(AuthContextKey)
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		user, ok := userValue.(*model.User)
		if !ok {
			response.Unauthorized(c, "Invalid user context")
			c.Abort()
			return
		}

		// If user is admin, allow access
		if user.IsAdmin() {
			c.Next()
			return
		}

		// If not admin, check if user owns the resource
		resourceUserID := getUserIDFromPath(c)
		if resourceUserID == 0 {
			response.BadRequest(c, "Invalid resource ID")
			c.Abort()
			return
		}

		// Check if user owns the resource
		if user.ID != resourceUserID {
			response.Forbidden(c, "Access denied: you can only access your own resources")
			c.Abort()
			return
		}

		c.Next()
	})
}