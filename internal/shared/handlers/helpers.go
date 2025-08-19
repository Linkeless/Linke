package handlers

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"linke/internal/shared/constants"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"
)

// BindAndValidate binds and validates JSON request with generic type
func BindAndValidate[T any](c *gin.Context) (*T, error) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.ErrInvalidRequestData)
		return nil, err
	}
	return &req, nil
}

// BindAndValidateQuery binds and validates query parameters with generic type
func BindAndValidateQuery[T any](c *gin.Context) (*T, error) {
	var req T
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, constants.ErrInvalidRequestData)
		return nil, err
	}
	return &req, nil
}

// HandleServiceError handles service errors with consistent error responses
func HandleServiceError(c *gin.Context, err error, entity string) {
	if err == nil {
		return
	}

	// Log the error with context
	logger.Error(fmt.Sprintf("Failed to process %s", entity),
		logger.String("path", c.Request.URL.Path),
		logger.String("method", c.Request.Method),
		logger.ErrorField(err),
	)

	// Determine error type and respond accordingly
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "not found"):
		response.NotFound(c, fmt.Sprintf("%s not found", entity))
	case strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "already exists"):
		response.Conflict(c, fmt.Sprintf("%s already exists", entity))
	case strings.Contains(errStr, "unauthorized"):
		response.Unauthorized(c, "Unauthorized access")
	case strings.Contains(errStr, "forbidden"):
		response.Forbidden(c, "Access forbidden")
	case strings.Contains(errStr, "validation"):
		response.BadRequest(c, err.Error())
	default:
		response.InternalServerError(c, fmt.Sprintf("Failed to process %s", entity))
	}
}

// SendPaginatedResponse sends a paginated response with generic items
func SendPaginatedResponse[T any](c *gin.Context, items []T, total int64) {
	response.SendPaginatedResponse(c, items, total)
}

// SendListResponse sends a list response without pagination
func SendListResponse[T any](c *gin.Context, items []T) {
	response.OK(c, items)
}

// SendSingleResponse sends a single item response
func SendSingleResponse[T any](c *gin.Context, item T) {
	response.OK(c, item)
}

// SendCreatedResponse sends a created response with the new item
func SendCreatedResponse[T any](c *gin.Context, item T) {
	response.Created(c, item)
}

// SendNoContentResponse sends a no content response
func SendNoContentResponse(c *gin.Context) {
	response.NoContent(c)
}

// PageParams for page-based pagination
type PageParams struct {
	Page int `form:"page" binding:"omitempty,min=1"`
	Size int `form:"size" binding:"omitempty,min=1,max=100"`
}

// GetPageParams gets and validates page-based pagination parameters
func GetPageParams(c *gin.Context) PageParams {
	var params PageParams
	c.ShouldBindQuery(&params)

	// Set defaults
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Size == 0 {
		params.Size = 20
	}
	if params.Size > 100 {
		params.Size = 100
	}

	return params
}

// GetOffset calculates offset from page params
func (p PageParams) GetOffset() int {
	return (p.Page - 1) * p.Size
}

// ParseFilterParams parses common filter parameters
type FilterParams struct {
	Search  string `form:"search"`
	Status  string `form:"status"`
	Sort    string `form:"sort"`
	OrderBy string `form:"order_by"`
	From    string `form:"from"`
	To      string `form:"to"`
}

// GetFilterParams gets filter parameters from query
func GetFilterParams(c *gin.Context) FilterParams {
	var params FilterParams
	c.ShouldBindQuery(&params)

	// Set default ordering if not specified
	if params.OrderBy == "" {
		params.OrderBy = "created_at DESC"
	}

	return params
}

// ValidateRequiredParams validates that required parameters are present
func ValidateRequiredParams(c *gin.Context, params map[string]interface{}) bool {
	missingParams := []string{}

	for name, value := range params {
		if value == nil || value == "" || value == 0 {
			missingParams = append(missingParams, name)
		}
	}

	if len(missingParams) > 0 {
		response.BadRequest(c, fmt.Sprintf("Missing required parameters: %s", strings.Join(missingParams, ", ")))
		return false
	}

	return true
}

// ExtractUserID extracts user ID from context (set by auth middleware)
func ExtractUserID(c *gin.Context) (uint, bool) {
	user, exists := GetCurrentUser(c)
	if !exists {
		return 0, false
	}
	return user.ID, true
}

// ExtractUserRole extracts user role from context
func ExtractUserRole(c *gin.Context) (string, bool) {
	user, exists := GetCurrentUser(c)
	if !exists {
		return "", false
	}
	return user.Role, true
}

// CheckUserRole checks if user has the required role
func CheckUserRole(c *gin.Context, requiredRole string) bool {
	role, exists := ExtractUserRole(c)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return false
	}

	if role != requiredRole {
		response.Forbidden(c, "Insufficient permissions")
		return false
	}

	return true
}

// HandlePanic recovers from panic and sends error response
func HandlePanic(c *gin.Context) {
	if r := recover(); r != nil {
		logger.Error("Panic recovered",
			logger.String("path", c.Request.URL.Path),
			logger.String("method", c.Request.Method),
			logger.Any("panic", r),
		)
		response.InternalServerError(c, "An unexpected error occurred")
	}
}
