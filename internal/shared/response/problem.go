package response

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ProblemType defines common problem types following RFC 9457
const (
	ProblemTypeBadRequest          = "/problems/bad-request"
	ProblemTypeUnauthorized        = "/problems/unauthorized" 
	ProblemTypeForbidden           = "/problems/forbidden"
	ProblemTypeNotFound            = "/problems/not-found"
	ProblemTypeMethodNotAllowed    = "/problems/method-not-allowed"
	ProblemTypeConflict            = "/problems/conflict"
	ProblemTypeUnprocessableEntity = "/problems/unprocessable-entity"
	ProblemTypeInternalServerError = "/problems/internal-server-error"
	ProblemTypeNotImplemented      = "/problems/not-implemented"
	ProblemTypeServiceUnavailable  = "/problems/service-unavailable"
	
	// Business-specific problem types
	ProblemTypeValidationFailed    = "/problems/validation-failed"
	ProblemTypeResourceExists      = "/problems/resource-exists"
	ProblemTypeResourceNotFound    = "/problems/resource-not-found"
	ProblemTypeInsufficientQuota   = "/problems/insufficient-quota"
	ProblemTypeRateLimitExceeded   = "/problems/rate-limit-exceeded"
	ProblemTypeAuthenticationFailed = "/problems/authentication-failed"
	ProblemTypeTokenExpired        = "/problems/token-expired"
	ProblemTypePermissionDenied    = "/problems/permission-denied"
)

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ProblemDetail creates detailed problem responses with proper extensions
func ProblemDetail(c *gin.Context, status int, problemType, title, detail string, extensions ...map[string]any) {
	problem := ProblemJSON{
		Type:   problemType,
		Title:  title,
		Status: status,
		Detail: detail,
		Instance: c.Request.URL.Path,
	}
	
	if len(extensions) > 0 {
		problem.Extensions = extensions[0]
	}
	
	c.Header("Content-Type", "application/problem+json")
	c.JSON(status, problem)
}

// ValidationFailed sends a validation failed problem response with field errors
func ValidationFailed(c *gin.Context, errors []ValidationError) {
	extensions := map[string]any{
		"invalid_params": errors,
	}
	
	detail := fmt.Sprintf("The request contained %d validation error(s)", len(errors))
	if len(errors) == 1 {
		detail = fmt.Sprintf("The request contained a validation error: %s", errors[0].Message)
	}
	
	ProblemDetail(c, http.StatusUnprocessableEntity, ProblemTypeValidationFailed, 
		"Validation Failed", detail, extensions)
}

// ResourceNotFound sends a resource not found problem response
func ResourceNotFound(c *gin.Context, resourceType, resourceID string) {
	detail := fmt.Sprintf("The %s with identifier '%s' was not found", resourceType, resourceID)
	extensions := map[string]any{
		"resource_type": resourceType,
		"resource_id":   resourceID,
	}
	
	ProblemDetail(c, http.StatusNotFound, ProblemTypeResourceNotFound, 
		"Resource Not Found", detail, extensions)
}

// ResourceAlreadyExists sends a resource conflict problem response
func ResourceAlreadyExists(c *gin.Context, resourceType, field, value string) {
	detail := fmt.Sprintf("A %s with %s '%s' already exists", resourceType, field, value)
	extensions := map[string]any{
		"resource_type": resourceType,
		"conflicting_field": field,
		"conflicting_value": value,
	}
	
	ProblemDetail(c, http.StatusConflict, ProblemTypeResourceExists, 
		"Resource Already Exists", detail, extensions)
}

// AuthenticationRequired sends an authentication required problem response
func AuthenticationRequired(c *gin.Context, scheme string) {
	detail := "Valid authentication credentials are required to access this resource"
	extensions := map[string]any{
		"authentication_scheme": scheme,
		"www_authenticate": fmt.Sprintf("Bearer realm=\"%s\"", scheme),
	}
	
	c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"%s\"", scheme))
	ProblemDetail(c, http.StatusUnauthorized, ProblemTypeUnauthorized, 
		"Authentication Required", detail, extensions)
}

// TokenExpired sends a token expired problem response
func TokenExpired(c *gin.Context) {
	detail := "The provided authentication token has expired"
	extensions := map[string]any{
		"error": "token_expired",
		"www_authenticate": "Bearer error=\"invalid_token\", error_description=\"Token has expired\"",
	}
	
	c.Header("WWW-Authenticate", "Bearer error=\"invalid_token\", error_description=\"Token has expired\"")
	ProblemDetail(c, http.StatusUnauthorized, ProblemTypeTokenExpired, 
		"Token Expired", detail, extensions)
}

// InsufficientPrivileges sends an insufficient privileges problem response
func InsufficientPrivileges(c *gin.Context, requiredRole string) {
	detail := fmt.Sprintf("Access denied. This operation requires '%s' privileges", requiredRole)
	extensions := map[string]any{
		"required_role": requiredRole,
	}
	
	ProblemDetail(c, http.StatusForbidden, ProblemTypePermissionDenied, 
		"Insufficient Privileges", detail, extensions)
}

// RateLimitExceeded sends a rate limit exceeded problem response
func RateLimitExceeded(c *gin.Context, limit int, resetTime int64) {
	detail := fmt.Sprintf("Rate limit of %d requests exceeded. Try again later", limit)
	extensions := map[string]any{
		"rate_limit": limit,
		"reset_time": resetTime,
	}
	
	c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
	c.Header("X-RateLimit-Remaining", "0")
	c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))
	c.Header("Retry-After", fmt.Sprintf("%d", resetTime))
	
	ProblemDetail(c, http.StatusTooManyRequests, ProblemTypeRateLimitExceeded, 
		"Rate Limit Exceeded", detail, extensions)
}

// InsufficientQuota sends an insufficient quota problem response
func InsufficientQuota(c *gin.Context, quotaType string, used, limit int64) {
	detail := fmt.Sprintf("Insufficient %s quota. Used: %d/%d", quotaType, used, limit)
	extensions := map[string]any{
		"quota_type": quotaType,
		"quota_used": used,
		"quota_limit": limit,
	}
	
	ProblemDetail(c, http.StatusPaymentRequired, ProblemTypeInsufficientQuota, 
		"Insufficient Quota", detail, extensions)
}

// MethodNotAllowed sends a method not allowed problem response
func MethodNotAllowed(c *gin.Context, allowedMethods []string) {
	detail := fmt.Sprintf("The %s method is not allowed for this resource", c.Request.Method)
	extensions := map[string]any{
		"method": c.Request.Method,
		"allowed_methods": allowedMethods,
	}
	
	c.Header("Allow", strings.Join(allowedMethods, ", "))
	ProblemDetail(c, http.StatusMethodNotAllowed, ProblemTypeMethodNotAllowed, 
		"Method Not Allowed", detail, extensions)
}

// MaintenanceMode sends a service unavailable problem response
func MaintenanceMode(c *gin.Context, retryAfter int) {
	detail := "The service is temporarily unavailable due to maintenance"
	extensions := map[string]any{
		"retry_after": retryAfter,
		"maintenance": true,
	}
	
	if retryAfter > 0 {
		c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
	}
	
	ProblemDetail(c, http.StatusServiceUnavailable, ProblemTypeServiceUnavailable, 
		"Service Unavailable", detail, extensions)
}

// HandleBusinessError handles domain-specific business errors
func HandleBusinessError(c *gin.Context, err error) {
	switch {
	case strings.Contains(err.Error(), "not found"):
		NotFound(c, err.Error())
	case strings.Contains(err.Error(), "already exists"):
		Conflict(c, err.Error())
	case strings.Contains(err.Error(), "validation"):
		UnprocessableEntity(c, err.Error())
	case strings.Contains(err.Error(), "unauthorized"):
		Unauthorized(c, err.Error())
	case strings.Contains(err.Error(), "forbidden"):
		Forbidden(c, err.Error())
	case strings.Contains(err.Error(), "quota"):
		InsufficientQuota(c, "general", 0, 0)
	case strings.Contains(err.Error(), "rate limit"):
		RateLimitExceeded(c, 100, 3600)
	default:
		InternalServerError(c, "An unexpected error occurred")
	}
}