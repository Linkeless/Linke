package shared

import (
	"errors"
	"strconv"
	"strings"

	"linke/internal/response"

	"github.com/gin-gonic/gin"
)

// TicketValidator provides validation utilities for ticket operations
type TicketValidator struct{}

// NewTicketValidator creates a new ticket validator
func NewTicketValidator() *TicketValidator {
	return &TicketValidator{}
}

// ValidateTicketID validates and parses ticket ID from URL parameter
func (v *TicketValidator) ValidateTicketID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID", err.Error())
		return 0, errors.New("invalid ticket ID")
	}
	return uint(id), nil
}

// ValidateTicketNumber validates ticket number
func (v *TicketValidator) ValidateTicketNumber(ticketNo string) error {
	if strings.TrimSpace(ticketNo) == "" {
		return errors.New("ticket number is required")
	}
	return nil
}

// ValidateTicketStatus validates ticket status value
func (v *TicketValidator) ValidateTicketStatus(status string) error {
	validStatuses := map[string]bool{
		"open":        true,
		"in_progress": true,
		"pending":     true,
		"resolved":    true,
		"closed":      true,
	}

	if !validStatuses[status] {
		return errors.New("invalid status value, must be one of: open, in_progress, pending, resolved, closed")
	}
	return nil
}

// ValidateTicketPriority validates ticket priority value
func (v *TicketValidator) ValidateTicketPriority(priority string) error {
	validPriorities := map[string]bool{
		"low":      true,
		"normal":   true,
		"high":     true,
		"urgent":   true,
		"critical": true,
	}

	if !validPriorities[priority] {
		return errors.New("invalid priority value, must be one of: low, normal, high, urgent, critical")
	}
	return nil
}

// ValidateTicketCategory validates ticket category value
func (v *TicketValidator) ValidateTicketCategory(category string) error {
	validCategories := map[string]bool{
		"general":      true,
		"technical":    true,
		"billing":      true,
		"account":      true,
		"feature":      true,
		"bug":          true,
		"subscription": true,
		"payment":      true,
	}

	if !validCategories[category] {
		return errors.New("invalid category value")
	}
	return nil
}

// ValidateMessageType validates ticket message type
func (v *TicketValidator) ValidateMessageType(messageType string) error {
	validTypes := map[string]bool{
		"user":   true,
		"admin":  true,
		"system": true,
	}

	if !validTypes[messageType] {
		return errors.New("invalid message type, must be one of: user, admin, system")
	}
	return nil
}

// ValidatePaginationParams validates and returns pagination parameters
func (v *TicketValidator) ValidatePaginationParams(c *gin.Context) (int, int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	return page, limit, offset
}

// ValidateSearchQuery validates search query parameter
func (v *TicketValidator) ValidateSearchQuery(search string) error {
	// Search is optional, so empty string is valid
	if len(search) > 500 {
		return errors.New("search query must be less than 500 characters")
	}
	return nil
}

// ValidateUserID validates user ID parameter
func (v *TicketValidator) ValidateUserID(userIDStr string) (uint, error) {
	if userIDStr == "" {
		return 0, nil // Optional parameter
	}
	
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, errors.New("invalid user ID")
	}
	
	return uint(userID), nil
}

// ValidateAssignedToID validates assigned admin ID parameter
func (v *TicketValidator) ValidateAssignedToID(assignedToIDStr string) (uint, error) {
	if assignedToIDStr == "" {
		return 0, nil // Optional parameter
	}
	
	assignedToID, err := strconv.ParseUint(assignedToIDStr, 10, 32)
	if err != nil {
		return 0, errors.New("invalid assigned admin ID")
	}
	
	return uint(assignedToID), nil
}

// ValidateAssignmentRequest validates ticket assignment request
func (v *TicketValidator) ValidateAssignmentRequest(adminID uint) error {
	if adminID == 0 {
		return errors.New("admin ID is required for assignment")
	}
	return nil
}

// ValidateResolutionMessage validates resolution message
func (v *TicketValidator) ValidateResolutionMessage(message string) error {
	if strings.TrimSpace(message) == "" {
		return errors.New("resolution message is required")
	}
	if len(message) > 2000 {
		return errors.New("resolution message must be less than 2000 characters")
	}
	return nil
}

// ValidateTicketMessage validates ticket message content
func (v *TicketValidator) ValidateTicketMessage(content string) error {
	if strings.TrimSpace(content) == "" {
		return errors.New("message content is required")
	}
	if len(content) > 5000 {
		return errors.New("message content must be less than 5000 characters")
	}
	return nil
}

// ValidateSortParams validates sort parameters
func (v *TicketValidator) ValidateSortParams(c *gin.Context) (string, string) {
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	
	validSortFields := map[string]bool{
		"created_at":     true,
		"updated_at":     true,
		"title":          true,
		"status":         true,
		"priority":       true,
		"category":       true,
		"ticket_number":  true,
	}
	
	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}
	
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}
	
	return sortBy, sortOrder
}