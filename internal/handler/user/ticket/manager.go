package ticket

import (
	ticketmanagement "linke/internal/handler/user/ticket/management"
	ticketoperation "linke/internal/handler/user/ticket/operation"
	ticketquery "linke/internal/handler/user/ticket/query"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// UserTicketManager manages all user ticket-related operations with modular structure
type UserTicketManager struct {
	// Sub-modules
	Management *ticketmanagement.TicketManagementHandler
	Query      *ticketquery.TicketQueryHandler
	Messages   *ticketoperation.TicketMessageHandler
}

// NewUserTicketManager creates a new user ticket manager
func NewUserTicketManager(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *UserTicketManager {
	return &UserTicketManager{
		Management: ticketmanagement.NewTicketManagementHandler(ticketService, ticketMessageService),
		Query:      ticketquery.NewTicketQueryHandler(ticketService, ticketMessageService),
		Messages:   ticketoperation.NewTicketMessageHandler(ticketService, ticketMessageService),
	}
}

// ============= Compatibility Methods =============
// These methods provide backward compatibility with existing code

// CreateTicket provides backward compatibility for ticket creation
func (m *UserTicketManager) CreateTicket(c *gin.Context) {
	m.Management.CreateTicket(c)
}

// GetMyTickets provides backward compatibility for getting user's tickets
func (m *UserTicketManager) GetMyTickets(c *gin.Context) {
	m.Query.GetMyTickets(c)
}

// GetTicket provides backward compatibility for getting ticket by ID
func (m *UserTicketManager) GetTicket(c *gin.Context) {
	m.Management.GetTicket(c)
}

// AddTicketMessage provides backward compatibility for adding ticket message
func (m *UserTicketManager) AddTicketMessage(c *gin.Context) {
	m.Messages.AddTicketMessage(c)
}

// GetTicketMessages provides backward compatibility for getting ticket messages
func (m *UserTicketManager) GetTicketMessages(c *gin.Context) {
	m.Messages.GetTicketMessages(c)
}

// GetTicketByNumber provides backward compatibility for getting ticket by number
func (m *UserTicketManager) GetTicketByNumber(c *gin.Context) {
	m.Query.GetTicketByNumber(c)
}

// SetupUserTicketRoutes sets up routes for user ticket operations
func (m *UserTicketManager) SetupUserTicketRoutes(userGroup *gin.RouterGroup) {
	// Note: Authentication middleware is already applied by parent group
	// userGroup.Use(middleware.AuthMiddleware(authService))

	// Ticket routes
	ticketGroup := userGroup.Group("/tickets")
	{
		ticketGroup.POST("", m.CreateTicket)
		ticketGroup.GET("", m.GetMyTickets)
		ticketGroup.GET("/:id", m.GetTicket)
		ticketGroup.GET("/number/:ticket_no", m.GetTicketByNumber)
		ticketGroup.POST("/:id/messages", m.AddTicketMessage)
		ticketGroup.GET("/:id/messages", m.GetTicketMessages)
	}
}