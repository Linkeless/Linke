package errors

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// ConvertServiceError converts service layer errors to appropriate business errors
func ConvertServiceError(err error, resource string, resourceID interface{}) error {
	if err == nil {
		return nil
	}

	// Handle GORM errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		switch resource {
		case "ticket":
			if id, ok := resourceID.(uint); ok {
				return NewTicketNotFound(id)
			}
		case "message":
			if id, ok := resourceID.(uint); ok {
				return NewTicketMessageNotFound(id)
			}
		case "user":
			if id, ok := resourceID.(uint); ok {
				return NewUserNotFound(id)
			}
		}
		return Newf(ErrCodeNotFound, "%s not found", resource)
	}

	// If it's already a business error, return as-is
	if IsBusinessError(err) {
		return err
	}

	// For unknown errors, wrap as internal error
	return Wrap(err, ErrCodeInternal, "Internal service error")
}

// HandlePermissionError creates a standardized permission error
func HandlePermissionError(userID, resourceID uint, resource, action string) error {
	switch resource {
	case "ticket":
		return NewTicketForbidden(resourceID, userID, action)
	default:
		return Newf(ErrCodeForbidden, "User %d is not allowed to %s %s %d", userID, action, resource, resourceID)
	}
}

// ValidateTicketState validates if a ticket state allows certain operations
func ValidateTicketState(ticketID uint, currentState, action string) error {
	switch action {
	case "add_message":
		if currentState == "closed" {
			return NewInvalidTicketState(ticketID, currentState, action, "Cannot add messages to closed tickets")
		}
	case "close":
		if currentState == "closed" {
			return NewInvalidTicketState(ticketID, currentState, action, "Ticket is already closed")
		}
	case "reopen":
		if currentState != "closed" {
			return NewInvalidTicketState(ticketID, currentState, action, "Only closed tickets can be reopened")
		}
	case "assign":
		if currentState == "closed" {
			return NewInvalidTicketState(ticketID, currentState, action, "Cannot assign closed tickets")
		}
	case "resolve":
		if currentState == "closed" || currentState == "resolved" {
			return NewInvalidTicketState(ticketID, currentState, action, "Ticket is already resolved or closed")
		}
	}
	return nil
}

// WrapWithContext wraps an error with additional context information
func WrapWithContext(err error, context map[string]interface{}) error {
	if err == nil {
		return nil
	}

	if be, ok := AsBusinessError(err); ok {
		// Add context to existing business error
		if context["detail"] != nil {
			if detail, ok := context["detail"].(string); ok && be.Detail == "" {
				be.Detail = detail
			}
		}
		if context["field"] != nil {
			if field, ok := context["field"].(string); ok && be.Field == "" {
				be.Field = field
			}
		}
		if context["resource"] != nil {
			if resource, ok := context["resource"].(string); ok && be.Resource == "" {
				be.Resource = resource
			}
		}
		if context["resource_id"] != nil {
			if resourceID, ok := context["resource_id"].(string); ok && be.ResourceID == "" {
				be.ResourceID = resourceID
			}
		}
		return be
	}

	// Create new business error with context
	detail := ""
	if context["detail"] != nil {
		if d, ok := context["detail"].(string); ok {
			detail = d
		}
	}

	be := &BusinessError{
		Code:    ErrCodeInternal,
		Message: err.Error(),
		Detail:  detail,
		Cause:   err,
	}

	if context["field"] != nil {
		if field, ok := context["field"].(string); ok {
			be.Field = field
		}
	}
	if context["resource"] != nil {
		if resource, ok := context["resource"].(string); ok {
			be.Resource = resource
		}
	}
	if context["resource_id"] != nil {
		if resourceID, ok := context["resource_id"].(string); ok {
			be.ResourceID = resourceID
		}
	}

	return be
}

// LogError logs an error with appropriate level based on error type
func LogError(err error, logger interface{}) {
	if err == nil {
		return
	}

	// This is a placeholder for logger interface
	// In practice, you would use your actual logger implementation
	fmt.Printf("Error: %v\n", err)
}

// SafeErrorMessage returns a safe error message for client consumption
func SafeErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	// For business errors, return the message as-is
	if be, ok := AsBusinessError(err); ok {
		return be.Message
	}

	// For domain-specific errors, return their messages
	switch err.(type) {
	case *TicketNotFoundError, *TicketMessageNotFoundError, *UserNotFoundError:
		return err.Error()
	case *TicketForbiddenError:
		return "You don't have permission to perform this action"
	case *InvalidTicketStateError:
		return err.Error()
	default:
		// For unknown errors, return a generic message
		return "An internal error occurred"
	}
}

// GetHTTPStatusCode returns the appropriate HTTP status code for an error
func GetHTTPStatusCode(err error) int {
	if err == nil {
		return 200
	}

	// Check for business errors first
	if be, ok := AsBusinessError(err); ok {
		return be.HTTPStatusCode()
	}

	// Check for domain-specific errors
	switch err.(type) {
	case *TicketNotFoundError, *TicketMessageNotFoundError, *UserNotFoundError:
		return 404
	case *TicketForbiddenError:
		return 403
	case *InvalidTicketStateError:
		return 400
	default:
		return 500
	}
}