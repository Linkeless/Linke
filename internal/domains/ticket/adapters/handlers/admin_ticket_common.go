package handlers

import (
	"fmt"

	ticketInterfaces "linke/internal/domains/ticket/usecases/interfaces"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/handlers"

	"github.com/gin-gonic/gin"
)

// AdminTicketHandlerBase provides base functionality for all admin ticket handlers
type AdminTicketHandlerBase struct {
	ticketService        ticketInterfaces.TicketService
	ticketMessageService ticketInterfaces.TicketMessageService
	userService          userInterfaces.UserService
}

// NewAdminTicketHandlerBase creates a new admin ticket handler base
func NewAdminTicketHandlerBase(
	ticketService ticketInterfaces.TicketService,
	ticketMessageService ticketInterfaces.TicketMessageService,
	userService userInterfaces.UserService,
) *AdminTicketHandlerBase {
	return &AdminTicketHandlerBase{
		ticketService:        ticketService,
		ticketMessageService: ticketMessageService,
		userService:          userService,
	}
}

// getAdminUserFromContext extracts and validates admin user from request context
func (h *AdminTicketHandlerBase) getAdminUserFromContext(c *gin.Context) (uint, error) {
	user, ok := handlers.GetCurrentUser(c)
	if !ok {
		return 0, fmt.Errorf("user not found in request context")
	}

	if user.Role != "admin" {
		return 0, fmt.Errorf("insufficient privileges: admin role required, got role: %s", user.Role)
	}

	return user.ID, nil
}
