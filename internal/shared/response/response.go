package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    any `json:"data,omitempty"`
}

// PaginationResponse represents pagination data structure
type PaginationResponse struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Items      any         `json:"items"`
	Pagination *PaginationResponse `json:"pagination"`
}

// ServerGroupResponseData represents server group data for responses
type ServerGroupResponseData struct {
	ID        uint   `json:"id" example:"1"`
	Name      string `json:"name" example:"Asia Pacific"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ErrorResponse represents an error response structure
type ErrorResponse struct {
	Error   string         `json:"error"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// SuccessResponse represents a success response structure
type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success sends a successful response
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage sends a successful response with custom message
func SuccessWithMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 created response
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, APIResponse{
		Code:    0,
		Message: "created successfully",
		Data:    data,
	})
}

// CreatedWithMessage sends a 201 created response with custom message
func CreatedWithMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, APIResponse{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// OK sends a 200 OK response with custom message and data
func OK(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// Error sends an error response
func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, APIResponse{
		Code:    code,
		Message: message,
	})
}

// BadRequest sends a 400 bad request response
func BadRequest(c *gin.Context, message string, details ...string) {
	msg := message
	if len(details) > 0 {
		msg = message + ": " + details[0]
	}
	Error(c, http.StatusBadRequest, 4000, msg)
}

// Unauthorized sends a 401 unauthorized response
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, 4001, message)
}

// Forbidden sends a 403 forbidden response
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, 4003, message)
}

// NotFound sends a 404 not found response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, 4004, message)
}

// Conflict sends a 409 conflict response
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, 4009, message)
}

// InternalServerError sends a 500 internal server error response
func InternalServerError(c *gin.Context, message string, details ...string) {
	msg := message
	if len(details) > 0 {
		msg = message + ": " + details[0]
	}
	Error(c, http.StatusInternalServerError, 5000, msg)
}

// ErrorJSON sends an error response with ErrorResponse structure
func ErrorJSON(c *gin.Context, httpStatus int, errorResponse ErrorResponse) {
	c.JSON(httpStatus, errorResponse)
}

// OKPaginated sends a paginated response with custom message
func OKPaginated(c *gin.Context, message string, data any, total int64, limit int, offset int) {
	response := struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    any `json:"data"`
		Total   int64       `json:"total"`
		Limit   int         `json:"limit"`
		Offset  int         `json:"offset"`
	}{
		Code:    0,
		Message: message,
		Data:    data,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}
	c.JSON(http.StatusOK, response)
}

// SuccessList sends a successful paginated list response
func SuccessList(c *gin.Context, items any, page int, limit int, total int64) {
	response := ListResponse{
		Items: items,
		Pagination: &PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}

	Success(c, response)
}

// SuccessListWithMessage sends a successful paginated list response with custom message
func SuccessListWithMessage(c *gin.Context, message string, items any, page int, limit int, total int64) {
	response := ListResponse{
		Items: items,
		Pagination: &PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}

	SuccessWithMessage(c, message, response)
}

// SuccessListWithExtra sends a successful paginated list response with additional data
func SuccessListWithExtra(c *gin.Context, message string, items any, page int, limit int, total int64, extra map[string]any) {
	response := map[string]any{
		"items": items,
		"pagination": &PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}

	// Add extra fields
	for key, value := range extra {
		response[key] = value
	}

	SuccessWithMessage(c, message, response)
}

// SuccessJSON sends a successful response with custom HTTP status
func SuccessJSON(c *gin.Context, httpStatus int, data any) {
	c.JSON(httpStatus, APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// HandleError is a helper function to handle errors with logging
func HandleError(logger any, c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Log the error if logger is provided
	if l, ok := logger.(interface {
		Error(msg string, fields ...any)
	}); ok {
		l.Error("Request error occurred", "error", err, "path", c.Request.URL.Path)
	}

	// Handle different error types
	switch err.Error() {
	case "record not found":
		NotFound(c, "Resource not found")
	case "unauthorized":
		Unauthorized(c, "Authentication required")
	case "forbidden":
		Forbidden(c, "Access denied")
	default:
		InternalServerError(c, "An error occurred while processing your request")
	}
}
