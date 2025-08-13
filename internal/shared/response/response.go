package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ProblemJSON represents RFC 9457 Problem Details for HTTP APIs
type ProblemJSON struct {
	Type       string         `json:"type"`               // URI reference that identifies the problem type
	Title      string         `json:"title"`              // Short, human-readable summary
	Status     int            `json:"status"`             // HTTP status code
	Detail     string         `json:"detail,omitempty"`   // Human-readable explanation
	Instance   string         `json:"instance,omitempty"` // URI reference that identifies the specific occurrence
	Extensions map[string]any `json:"-"`                  // Additional problem-specific fields
}

// MarshalJSON implements custom JSON marshaling to include extensions at the root level
func (p ProblemJSON) MarshalJSON() ([]byte, error) {
	// Convert to map to include extensions
	result := make(map[string]any)
	result["type"] = p.Type
	result["title"] = p.Title
	result["status"] = p.Status
	if p.Detail != "" {
		result["detail"] = p.Detail
	}
	if p.Instance != "" {
		result["instance"] = p.Instance
	}

	// Add extensions
	for k, v := range p.Extensions {
		result[k] = v
	}

	return json.Marshal(result)
}

// HALCollection represents a HAL-style collection response
type HALCollection struct {
	Embedded map[string]any `json:"_embedded"`
	Links    HALLinks       `json:"_links"`
	Page     PageInfo       `json:"page,omitempty"`
	Total    int64          `json:"total,omitempty"`
}

// HALResource represents a HAL-style resource with links
type HALResource struct {
	Links HALLinks `json:"_links,omitempty"`
}

// HALLinks represents HAL navigation links
type HALLinks struct {
	Self  *HALLink `json:"self,omitempty"`
	First *HALLink `json:"first,omitempty"`
	Prev  *HALLink `json:"prev,omitempty"`
	Next  *HALLink `json:"next,omitempty"`
	Last  *HALLink `json:"last,omitempty"`
}

// HALLink represents a single HAL link
type HALLink struct {
	Href      string `json:"href"`
	Templated bool   `json:"templated,omitempty"`
	Type      string `json:"type,omitempty"`
	Title     string `json:"title,omitempty"`
}

// PageInfo represents pagination metadata
type PageInfo struct {
	Size          int   `json:"size"`          // Items per page
	TotalElements int64 `json:"totalElements"` // Total number of elements
	TotalPages    int   `json:"totalPages"`    // Total number of pages
	Number        int   `json:"number"`        // Current page number (0-based)
}

// ServerGroupResponseData represents server group data for responses (kept for compatibility)
type ServerGroupResponseData struct {
	ID        uint   `json:"id" example:"1"`
	Name      string `json:"name" example:"Asia Pacific"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// Direct response functions - return resources directly without wrapping

// OK sends resource data directly with 200 status
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// Created sends resource data directly with 201 status
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}

// NoContent sends 204 No Content (for successful operations with no response body)
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Accepted sends 202 Accepted (for async operations)
func Accepted(c *gin.Context, data any) {
	if data != nil {
		c.JSON(http.StatusAccepted, data)
	} else {
		c.Status(http.StatusAccepted)
	}
}

// Collection sends a HAL-style collection response
func Collection(c *gin.Context, items any, links HALLinks, page *PageInfo, total int64) {
	response := HALCollection{
		Embedded: map[string]any{
			"items": items,
		},
		Links: links,
		Total: total,
	}

	if page != nil {
		response.Page = *page
	}

	c.JSON(http.StatusOK, response)
}

// Resource sends a HAL-style resource response with optional links
func Resource(c *gin.Context, data any, links ...HALLinks) {
	if len(links) > 0 && links[0].Self != nil {
		// If data is a map, add _links to it
		if dataMap, ok := data.(map[string]any); ok {
			dataMap["_links"] = links[0]
			c.JSON(http.StatusOK, dataMap)
			return
		}
		// For other types, wrap with links
		response := struct {
			Data  any      `json:",inline"`
			Links HALLinks `json:"_links"`
		}{
			Data:  data,
			Links: links[0],
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// No links, send data directly
	c.JSON(http.StatusOK, data)
}

// Problem response functions following RFC 9457

// Problem sends a RFC 9457 Problem JSON response
func Problem(c *gin.Context, status int, problemType, title, detail string, extensions ...map[string]any) {
	problem := ProblemJSON{
		Type:   problemType,
		Title:  title,
		Status: status,
		Detail: detail,
	}

	if len(extensions) > 0 {
		problem.Extensions = extensions[0]
	}

	c.Header("Content-Type", "application/problem+json")
	c.JSON(status, problem)
}

// BadRequest sends a 400 Bad Request problem response
func BadRequest(c *gin.Context, detail string, extensions ...map[string]any) {
	Problem(c, http.StatusBadRequest, "/problems/bad-request", "Bad Request", detail, extensions...)
}

// Unauthorized sends a 401 Unauthorized problem response
func Unauthorized(c *gin.Context, detail string) {
	Problem(c, http.StatusUnauthorized, "/problems/unauthorized", "Unauthorized", detail)
}

// Forbidden sends a 403 Forbidden problem response
func Forbidden(c *gin.Context, detail string) {
	Problem(c, http.StatusForbidden, "/problems/forbidden", "Forbidden", detail)
}

// NotFound sends a 404 Not Found problem response
func NotFound(c *gin.Context, detail string) {
	Problem(c, http.StatusNotFound, "/problems/not-found", "Not Found", detail)
}

// Conflict sends a 409 Conflict problem response
func Conflict(c *gin.Context, detail string, extensions ...map[string]any) {
	Problem(c, http.StatusConflict, "/problems/conflict", "Conflict", detail, extensions...)
}

// UnprocessableEntity sends a 422 Unprocessable Entity problem response
func UnprocessableEntity(c *gin.Context, detail string, extensions ...map[string]any) {
	Problem(c, http.StatusUnprocessableEntity, "/problems/unprocessable-entity", "Unprocessable Entity", detail, extensions...)
}

// InternalServerError sends a 500 Internal Server Error problem response
func InternalServerError(c *gin.Context, detail string) {
	Problem(c, http.StatusInternalServerError, "/problems/internal-server-error", "Internal Server Error", detail)
}

// NotImplemented sends a 501 Not Implemented problem response
func NotImplemented(c *gin.Context, detail string) {
	Problem(c, http.StatusNotImplemented, "/problems/not-implemented", "Not Implemented", detail)
}

// ServiceUnavailable sends a 503 Service Unavailable problem response
func ServiceUnavailable(c *gin.Context, detail string) {
	Problem(c, http.StatusServiceUnavailable, "/problems/service-unavailable", "Service Unavailable", detail)
}

// Legacy compatibility functions (deprecated - to be removed)

// Success - DEPRECATED: Use OK() instead
func Success(c *gin.Context, data any) {
	OK(c, data)
}

// SuccessWithMessage - DEPRECATED: HTTP responses should not have wrapper messages
func SuccessWithMessage(c *gin.Context, message string, data any) {
	// In RESTful APIs, success is indicated by HTTP status, not wrapper messages
	OK(c, data)
}

// CreatedWithMessage - DEPRECATED: Use Created() instead
func CreatedWithMessage(c *gin.Context, message string, data any) {
	Created(c, data)
}

// Error - DEPRECATED: Use specific Problem functions instead
func Error(c *gin.Context, httpStatus int, code int, message string) {
	Problem(c, httpStatus, fmt.Sprintf("/problems/error-%d", code), http.StatusText(httpStatus), message)
}

// ErrorJSON - DEPRECATED: Use Problem instead
func ErrorJSON(c *gin.Context, httpStatus int, errorResponse any) {
	c.JSON(httpStatus, errorResponse)
}

// SuccessJSON - DEPRECATED: Use OK instead
func SuccessJSON(c *gin.Context, httpStatus int, data any) {
	c.JSON(httpStatus, data)
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

	// Handle different error types with Problem JSON
	switch err.Error() {
	case "record not found":
		NotFound(c, "The requested resource was not found")
	case "unauthorized":
		Unauthorized(c, "Authentication credentials are required")
	case "forbidden":
		Forbidden(c, "Access to this resource is denied")
	default:
		if strings.Contains(err.Error(), "validation") {
			UnprocessableEntity(c, fmt.Sprintf("Validation failed: %s", err.Error()))
		} else {
			InternalServerError(c, "An internal server error occurred")
		}
	}
}

// BuildHALLinks creates HAL-style pagination links
func BuildHALLinks(baseURL string, page, limit int, total int64) HALLinks {
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	links := HALLinks{
		Self: &HALLink{
			Href: fmt.Sprintf("%s?page=%d&limit=%d", baseURL, page, limit),
		},
		First: &HALLink{
			Href: fmt.Sprintf("%s?page=1&limit=%d", baseURL, limit),
		},
		Last: &HALLink{
			Href: fmt.Sprintf("%s?page=%d&limit=%d", baseURL, totalPages, limit),
		},
	}

	if page > 1 {
		links.Prev = &HALLink{
			Href: fmt.Sprintf("%s?page=%d&limit=%d", baseURL, page-1, limit),
		}
	}

	if page < totalPages {
		links.Next = &HALLink{
			Href: fmt.Sprintf("%s?page=%d&limit=%d", baseURL, page+1, limit),
		}
	}

	return links
}

// BuildHALLinksWithQuery creates HAL-style pagination links with query parameters
func BuildHALLinksWithQuery(baseURL string, page, limit int, total int64, query map[string]any) HALLinks {
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Build query string
	queryStr := buildQueryString(query)
	separator := "?"
	if strings.Contains(baseURL, "?") {
		separator = "&"
	}

	links := HALLinks{
		Self: &HALLink{
			Href: fmt.Sprintf("%s%spage=%d&limit=%d%s", baseURL, separator, page, limit, queryStr),
		},
		First: &HALLink{
			Href: fmt.Sprintf("%s%spage=1&limit=%d%s", baseURL, separator, limit, queryStr),
		},
		Last: &HALLink{
			Href: fmt.Sprintf("%s%spage=%d&limit=%d%s", baseURL, separator, totalPages, limit, queryStr),
		},
	}

	if page > 1 {
		links.Prev = &HALLink{
			Href: fmt.Sprintf("%s%spage=%d&limit=%d%s", baseURL, separator, page-1, limit, queryStr),
		}
	}

	if page < totalPages {
		links.Next = &HALLink{
			Href: fmt.Sprintf("%s%spage=%d&limit=%d%s", baseURL, separator, page+1, limit, queryStr),
		}
	}

	return links
}

// DEPRECATED: Legacy pagination functions for backward compatibility

// Paginated - DEPRECATED: Use Collection with HAL links instead
func Paginated(c *gin.Context, message string, data any, page int, limit int, total int64, baseURL string) {
	links := BuildHALLinks(baseURL, page, limit, total)
	pageInfo := &PageInfo{
		Size:          limit,
		TotalElements: total,
		TotalPages:    int((total + int64(limit) - 1) / int64(limit)),
		Number:        page - 1, // HAL uses 0-based page numbers
	}

	Collection(c, data, links, pageInfo, total)
}

// PaginatedWithQuery - DEPRECATED: Use Collection with HAL links instead
func PaginatedWithQuery(c *gin.Context, message string, data any, page int, limit int, total int64, baseURL string, query map[string]any) {
	links := BuildHALLinksWithQuery(baseURL, page, limit, total, query)
	pageInfo := &PageInfo{
		Size:          limit,
		TotalElements: total,
		TotalPages:    int((total + int64(limit) - 1) / int64(limit)),
		Number:        page - 1, // HAL uses 0-based page numbers
	}

	Collection(c, data, links, pageInfo, total)
}

// buildQueryString builds query string from map
func buildQueryString(query map[string]any) string {
	if len(query) == 0 {
		return ""
	}

	var params []string
	for key, value := range query {
		if value != nil && value != "" {
			params = append(params, fmt.Sprintf("%s=%v", key, value))
		}
	}

	if len(params) == 0 {
		return ""
	}

	return "&" + strings.Join(params, "&")
}

// getStringFromQuery extracts string value from query map
func getStringFromQuery(query map[string]any, key string) string {
	if val, exists := query[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getFiltersFromQuery extracts filter parameters from query map
func getFiltersFromQuery(query map[string]any) map[string]interface{} {
	filters := make(map[string]interface{})
	for k, v := range query {
		if k != "search" && k != "page" && k != "limit" {
			filters[k] = v
		}
	}
	if len(filters) == 0 {
		return nil
	}
	return filters
}

// SetStandardHeaders sets common RESTful headers
func SetStandardHeaders(c *gin.Context) {
	// Set cache control headers
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// Set content type to JSON by default
	c.Header("Content-Type", "application/json")
}

// SetETag sets ETag header for caching
func SetETag(c *gin.Context, etag string) {
	c.Header("ETag", fmt.Sprintf(`"%s"`, etag))
}

// SetLastModified sets Last-Modified header
func SetLastModified(c *gin.Context, lastModified time.Time) {
	c.Header("Last-Modified", lastModified.Format(http.TimeFormat))
}

// SetLocation sets Location header for created resources
func SetLocation(c *gin.Context, location string) {
	c.Header("Location", location)
}
