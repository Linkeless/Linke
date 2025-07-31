package ticket

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"

	"github.com/gin-gonic/gin"
)

// TicketValidator provides validation utilities for ticket handlers
type TicketValidator struct{}

// NewTicketValidator creates a new ticket validator
func NewTicketValidator() *TicketValidator {
	return &TicketValidator{}
}

// ValidatePaginationParams validates and parses pagination parameters
func (v *TicketValidator) ValidatePaginationParams(c *gin.Context) (limit, offset int, valid bool) {
	// Get query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset, true
}

// GetUserFromContext extracts and validates user from context
func (v *TicketValidator) GetUserFromContext(c *gin.Context) (*model.User, bool) {
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		logger.Error("User not found in context")
		response.Unauthorized(c, "User not found")
		return nil, false
	}

	currentUser, ok := userValue.(*model.User)
	if !ok {
		logger.Error("Invalid user context")
		response.Unauthorized(c, "Invalid user context")
		return nil, false
	}

	return currentUser, true
}

// ValidateTicketIDParam validates and parses ticket ID parameter from path
func (v *TicketValidator) ValidateTicketIDParam(c *gin.Context) (uint, bool) {
	ticketIDStr := c.Param("id")
	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		logger.Error("Invalid ticket ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid ticket ID", err.Error())
		return 0, false
	}
	return uint(ticketID), true
}

// ValidateTicketNumberParam validates ticket number parameter from path
func (v *TicketValidator) ValidateTicketNumberParam(c *gin.Context) (string, bool) {
	ticketNo := c.Param("ticket_no")
	if ticketNo == "" {
		logger.Error("Empty ticket number")
		response.BadRequest(c, "Ticket number is required")
		return "", false
	}
	return ticketNo, true
}

// CheckTicketOwnership checks if user owns the ticket
func (v *TicketValidator) CheckTicketOwnership(user *model.User, ticket *model.Ticket, c *gin.Context) bool {
	if ticket.UserID != user.ID {
		logger.Error("Access denied to ticket", 
			logger.Uint("ticket_id", ticket.ID),
			logger.Uint("user_id", user.ID),
			logger.Uint("owner_id", ticket.UserID))
		response.Forbidden(c, "Access denied: you can only view your own tickets")
		return false
	}
	return true
}