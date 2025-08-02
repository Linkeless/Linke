package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserIDFromPath extracts user ID from the path parameter "id"
// This helper function is commonly used with RequireAdminOrOwner middleware
func GetUserIDFromPath(c *gin.Context) uint {
	idStr := c.Param("id")
	if idStr == "" {
		return 0
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0
	}

	return uint(id)
}

// GetUserIDFromQuery extracts user ID from query parameter "user_id"
// This helper function can be used for filtering operations
func GetUserIDFromQuery(c *gin.Context) uint {
	idStr := c.Query("user_id")
	if idStr == "" {
		return 0
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0
	}

	return uint(id)
}