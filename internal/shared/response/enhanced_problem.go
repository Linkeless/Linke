package response

import (
	"fmt"
	"net/http"
	"time"

	"linke/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

// EnhancedProblemJSON creates a comprehensive Problem JSON response
func EnhancedProblemJSON(c *gin.Context, err error) {
	// Extract context information
	ctx := extractProblemContext(c)
	
	// Create base problem response
	problem := &ProblemJSONResponse{
		Timestamp: time.Now(),
		Instance:  c.Request.URL.Path,
		TraceID:   ctx.TraceID,
	}

	// Add request ID to extensions if available
	if ctx.RequestID != "" {
		if problem.Extensions == nil {
			problem.Extensions = make(map[string]interface{})
		}
		problem.Extensions["request_id"] = ctx.RequestID
	}

	// Add user context if available
	if ctx.UserID != "" {
		if problem.Extensions == nil {
			problem.Extensions = make(map[string]interface{})
		}
		problem.Extensions["user_id"] = ctx.UserID
	}

	// Add operation context if available
	if ctx.Operation != "" {
		if problem.Extensions == nil {
			problem.Extensions = make(map[string]interface{})
		}
		problem.Extensions["operation"] = ctx.Operation
	}

	// Handle different error types
	if err == nil {
		// Generic internal server error
		problem.Type = ProblemTypeInternalServerError
		problem.Title = "Internal Server Error"
		problem.Status = http.StatusInternalServerError
		problem.Detail = "An unexpected error occurred"
	} else {
		// Handle business errors
		if be, ok := errors.AsBusinessError(err); ok {
			handleBusinessErrorInProblem(problem, be, ctx)
		} else if ve, ok := errors.AsValidationError(err); ok {
			handleValidationErrorInProblem(problem, ve, ctx)
		} else if ce, ok := errors.AsContextualError(err); ok {
			handleContextualErrorInProblem(problem, ce, ctx)
		} else {
			// Handle generic errors
			handleGenericErrorInProblem(problem, err, ctx)
		}
	}

	// Set appropriate headers
	c.Header("Content-Type", "application/problem+json")
	
	// Add caching headers for client errors
	if problem.Status >= 400 && problem.Status < 500 {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	// Send response
	c.JSON(problem.Status, problem)
}

// handleBusinessErrorInProblem handles BusinessError types
func handleBusinessErrorInProblem(problem *ProblemJSONResponse, be *errors.BusinessError, ctx *ProblemContext) {
	problem.Status = be.HTTPStatusCode()
	problem.Title = getErrorTitle(be.Code)
	problem.Type = getErrorType(be.Code)
	
	if be.Detail != "" {
		problem.Detail = be.Detail
	} else {
		problem.Detail = be.Message
	}

	// Add business error context
	if problem.Extensions == nil {
		problem.Extensions = make(map[string]interface{})
	}
	
	problem.Extensions["error_code"] = string(be.Code)
	
	if be.Resource != "" {
		problem.Extensions["resource"] = be.Resource
	}
	
	if be.ResourceID != "" {
		problem.Extensions["resource_id"] = be.ResourceID
	}
	
	if be.Field != "" {
		problem.Extensions["field"] = be.Field
	}
}

// handleValidationErrorInProblem handles ValidationError types
func handleValidationErrorInProblem(problem *ProblemJSONResponse, ve *errors.ValidationError, ctx *ProblemContext) {
	problem.Status = http.StatusUnprocessableEntity
	problem.Title = "Validation Failed"
	problem.Type = ProblemTypeValidationFailed
	problem.Detail = ve.Error()
	
	// Add field errors
	if ve.HasFields() {
		problem.Errors = ve.GetFieldErrors()
	}

	// Add validation summary
	if problem.Extensions == nil {
		problem.Extensions = make(map[string]interface{})
	}
	
	problem.Extensions["error_count"] = len(ve.GetFieldErrors())
	problem.Extensions["validation_failed"] = true
}

// handleContextualErrorInProblem handles ContextualError types
func handleContextualErrorInProblem(problem *ProblemJSONResponse, ce *errors.ContextualError, ctx *ProblemContext) {
	// Handle the underlying business error
	handleBusinessErrorInProblem(problem, ce.BusinessError, ctx)
	
	// Add contextual information
	if ce.Context != nil {
		if problem.Extensions == nil {
			problem.Extensions = make(map[string]interface{})
		}

		if ce.Context.TraceID != "" && problem.TraceID == "" {
			problem.TraceID = ce.Context.TraceID
		}

		if ce.Context.Operation != "" {
			problem.Extensions["operation"] = ce.Context.Operation
		}

		if len(ce.Context.Metadata) > 0 {
			problem.Extensions["error_metadata"] = ce.Context.Metadata
		}

		if len(ce.Context.Annotations) > 0 {
			problem.Extensions["annotations"] = ce.Context.Annotations
		}

		if len(ce.Context.StackTrace) > 0 {
			// Only include stack trace in development/debug mode
			problem.Extensions["stack_trace"] = ce.Context.StackTrace
		}
	}
}

// handleGenericErrorInProblem handles generic error types
func handleGenericErrorInProblem(problem *ProblemJSONResponse, err error, ctx *ProblemContext) {
	problem.Status = http.StatusInternalServerError
	problem.Title = "Internal Server Error"
	problem.Type = ProblemTypeInternalServerError
	problem.Detail = "An unexpected error occurred"

	// Add error information for debugging
	if problem.Extensions == nil {
		problem.Extensions = make(map[string]interface{})
	}
	
	problem.Extensions["error_type"] = fmt.Sprintf("%T", err)
	problem.Extensions["original_error"] = err.Error()
}

// getErrorType maps error codes to problem types
func getErrorType(code errors.ErrorCode) string {
	switch code {
	case errors.ErrCodeNotFound:
		return ProblemTypeNotFound
	case errors.ErrCodeAlreadyExists, errors.ErrCodeConflict:
		return ProblemTypeConflict
	case errors.ErrCodeForbidden:
		return ProblemTypeForbidden
	case errors.ErrCodeUnauthorized:
		return ProblemTypeUnauthorized
	case errors.ErrCodeInvalidInput, errors.ErrCodeInvalidFormat, errors.ErrCodeInvalidState:
		return ProblemTypeBadRequest
	case errors.ErrCodeTimeout:
		return "/problems/timeout"
	case errors.ErrCodeUnavailable:
		return ProblemTypeServiceUnavailable
	case errors.ErrCodeBusinessRule:
		return ProblemTypeUnprocessableEntity
	case errors.ErrCodeQuotaExceeded:
		return ProblemTypeInsufficientQuota
	default:
		return ProblemTypeInternalServerError
	}
}

// getErrorTitle maps error codes to human-readable titles
func getErrorTitle(code errors.ErrorCode) string {
	switch code {
	case errors.ErrCodeNotFound:
		return "Not Found"
	case errors.ErrCodeAlreadyExists:
		return "Already Exists"
	case errors.ErrCodeConflict:
		return "Conflict"
	case errors.ErrCodeForbidden:
		return "Forbidden"
	case errors.ErrCodeUnauthorized:
		return "Unauthorized"
	case errors.ErrCodeInvalidInput:
		return "Invalid Input"
	case errors.ErrCodeInvalidFormat:
		return "Invalid Format"
	case errors.ErrCodeInvalidState:
		return "Invalid State"
	case errors.ErrCodeInternal:
		return "Internal Server Error"
	case errors.ErrCodeTimeout:
		return "Request Timeout"
	case errors.ErrCodeUnavailable:
		return "Service Unavailable"
	case errors.ErrCodeBusinessRule:
		return "Business Rule Violation"
	case errors.ErrCodeQuotaExceeded:
		return "Quota Exceeded"
	default:
		return "Error"
	}
}

// extractProblemContext extracts context information from Gin context
func extractProblemContext(c *gin.Context) *ProblemContext {
	ctx := &ProblemContext{
		Metadata: make(map[string]interface{}),
	}

	// Extract trace ID
	if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
		ctx.TraceID = traceID
	} else if traceID := c.GetHeader("Trace-Id"); traceID != "" {
		ctx.TraceID = traceID
	}

	// Extract request ID
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		ctx.RequestID = requestID
	} else if requestID := c.GetHeader("Request-Id"); requestID != "" {
		ctx.RequestID = requestID
	}

	// Extract user ID from context
	if userID, exists := c.Get("user_id"); exists {
		if uidStr, ok := userID.(string); ok {
			ctx.UserID = uidStr
		}
	}

	// Extract operation from context
	if operation, exists := c.Get("operation"); exists {
		if opStr, ok := operation.(string); ok {
			ctx.Operation = opStr
		}
	}

	return ctx
}

// BadRequestProblem creates a bad request problem response
func BadRequestProblem(c *gin.Context, detail string, extensions ...map[string]interface{}) {
	problem := &ProblemJSONResponse{
		Type:      ProblemTypeBadRequest,
		Title:     "Bad Request",
		Status:    http.StatusBadRequest,
		Detail:    detail,
		Instance:  c.Request.URL.Path,
		Timestamp: time.Now(),
	}

	if len(extensions) > 0 {
		problem.Extensions = extensions[0]
	}

	ctx := extractProblemContext(c)
	if ctx.TraceID != "" {
		problem.TraceID = ctx.TraceID
	}

	c.Header("Content-Type", "application/problem+json")
	c.JSON(problem.Status, problem)
}

// NotFoundProblem creates a not found problem response
func NotFoundProblem(c *gin.Context, resourceType, resourceID string) {
	problem := &ProblemJSONResponse{
		Type:      ProblemTypeNotFound,
		Title:     "Not Found",
		Status:    http.StatusNotFound,
		Detail:    fmt.Sprintf("The %s with ID '%s' was not found", resourceType, resourceID),
		Instance:  c.Request.URL.Path,
		Timestamp: time.Now(),
		Extensions: map[string]interface{}{
			"resource_type": resourceType,
			"resource_id":   resourceID,
		},
	}

	ctx := extractProblemContext(c)
	if ctx.TraceID != "" {
		problem.TraceID = ctx.TraceID
	}

	c.Header("Content-Type", "application/problem+json")
	c.JSON(problem.Status, problem)
}

// ConflictProblem creates a conflict problem response
func ConflictProblem(c *gin.Context, detail string, conflictingField, conflictingValue string) {
	problem := &ProblemJSONResponse{
		Type:      ProblemTypeConflict,
		Title:     "Conflict",
		Status:    http.StatusConflict,
		Detail:    detail,
		Instance:  c.Request.URL.Path,
		Timestamp: time.Now(),
		Extensions: map[string]interface{}{
			"conflicting_field": conflictingField,
			"conflicting_value": conflictingValue,
		},
	}

	ctx := extractProblemContext(c)
	if ctx.TraceID != "" {
		problem.TraceID = ctx.TraceID
	}

	c.Header("Content-Type", "application/problem+json")
	c.JSON(problem.Status, problem)
}

// UnauthorizedProblem creates an unauthorized problem response
func UnauthorizedProblem(c *gin.Context, scheme string) {
	problem := &ProblemJSONResponse{
		Type:      ProblemTypeUnauthorized,
		Title:     "Unauthorized",
		Status:    http.StatusUnauthorized,
		Detail:    "Valid authentication credentials are required",
		Instance:  c.Request.URL.Path,
		Timestamp: time.Now(),
		Extensions: map[string]interface{}{
			"authentication_scheme": scheme,
		},
	}

	ctx := extractProblemContext(c)
	if ctx.TraceID != "" {
		problem.TraceID = ctx.TraceID
	}

	c.Header("Content-Type", "application/problem+json")
	c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"%s\"", scheme))
	c.JSON(problem.Status, problem)
}

// ForbiddenProblem creates a forbidden problem response
func ForbiddenProblem(c *gin.Context, requiredPermission string) {
	problem := &ProblemJSONResponse{
		Type:      ProblemTypeForbidden,
		Title:     "Forbidden",
		Status:    http.StatusForbidden,
		Detail:    "You do not have permission to access this resource",
		Instance:  c.Request.URL.Path,
		Timestamp: time.Now(),
		Extensions: map[string]interface{}{
			"required_permission": requiredPermission,
		},
	}

	ctx := extractProblemContext(c)
	if ctx.TraceID != "" {
		problem.TraceID = ctx.TraceID
	}

	c.Header("Content-Type", "application/problem+json")
	c.JSON(problem.Status, problem)
}

// ValidationProblem creates a validation problem response
func ValidationProblem(c *gin.Context, fieldErrors []errors.FieldError) {
	problem := &ProblemJSONResponse{
		Type:      ProblemTypeValidationFailed,
		Title:     "Validation Failed",
		Status:    http.StatusUnprocessableEntity,
		Detail:    fmt.Sprintf("The request contains %d validation error(s)", len(fieldErrors)),
		Instance:  c.Request.URL.Path,
		Timestamp: time.Now(),
		Errors:    fieldErrors,
		Extensions: map[string]interface{}{
			"error_count": len(fieldErrors),
		},
	}

	ctx := extractProblemContext(c)
	if ctx.TraceID != "" {
		problem.TraceID = ctx.TraceID
	}

	c.Header("Content-Type", "application/problem+json")
	c.JSON(problem.Status, problem)
}

// InternalServerErrorProblem creates an internal server error problem response
func InternalServerErrorProblem(c *gin.Context, includeDetails bool) {
	problem := &ProblemJSONResponse{
		Type:      ProblemTypeInternalServerError,
		Title:     "Internal Server Error",
		Status:    http.StatusInternalServerError,
		Detail:    "An unexpected error occurred",
		Instance:  c.Request.URL.Path,
		Timestamp: time.Now(),
	}

	ctx := extractProblemContext(c)
	if ctx.TraceID != "" {
		problem.TraceID = ctx.TraceID
	}

	// Add debugging information if enabled
	if includeDetails {
		if problem.Extensions == nil {
			problem.Extensions = make(map[string]interface{})
		}
		problem.Extensions["debug"] = true
	}

	c.Header("Content-Type", "application/problem+json")
	c.JSON(problem.Status, problem)
}

// ServiceUnavailableProblem creates a service unavailable problem response
func ServiceUnavailableProblem(c *gin.Context, retryAfter int, maintenance bool) {
	detail := "The service is temporarily unavailable"
	if maintenance {
		detail = "The service is temporarily unavailable due to maintenance"
	}

	problem := &ProblemJSONResponse{
		Type:      ProblemTypeServiceUnavailable,
		Title:     "Service Unavailable",
		Status:    http.StatusServiceUnavailable,
		Detail:    detail,
		Instance:  c.Request.URL.Path,
		Timestamp: time.Now(),
		Extensions: map[string]interface{}{
			"maintenance": maintenance,
		},
	}

	if retryAfter > 0 {
		problem.Extensions["retry_after"] = retryAfter
		c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
	}

	ctx := extractProblemContext(c)
	if ctx.TraceID != "" {
		problem.TraceID = ctx.TraceID
	}

	c.Header("Content-Type", "application/problem+json")
	c.JSON(problem.Status, problem)
}