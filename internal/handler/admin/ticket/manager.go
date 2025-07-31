package ticket

import (
	"linke/internal/handler/admin/ticket/management"
	"linke/internal/handler/admin/ticket/operation"
	"linke/internal/handler/admin/ticket/query"
	"linke/internal/handler/admin/ticket/statistics"
	"linke/internal/handler/admin/ticket/status"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminTicketManager manages all ticket-related admin handlers
type AdminTicketManager struct {
	// Sub-handlers for different ticket management aspects
	Management *management.TicketCRUDHandler
	List       *query.TicketListHandler
	Detail     *query.TicketDetailHandler
	Actions    *operation.TicketActionHandler
	Messages   *operation.TicketMessageHandler
	Status     *status.TicketStatusHandler
	Stats      *statistics.TicketStatsHandler
}

// NewAdminTicketManager creates a new admin ticket manager with all sub-handlers
func NewAdminTicketManager(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *AdminTicketManager {
	return &AdminTicketManager{
		Management: management.NewTicketCRUDHandler(ticketService, ticketMessageService),
		List:       query.NewTicketListHandler(ticketService, ticketMessageService),
		Detail:     query.NewTicketDetailHandler(ticketService, ticketMessageService),
		Actions:    operation.NewTicketActionHandler(ticketService, ticketMessageService),
		Messages:   operation.NewTicketMessageHandler(ticketService, ticketMessageService),
		Status:     status.NewTicketStatusHandler(ticketService, ticketMessageService),
		Stats:      statistics.NewTicketStatsHandler(ticketService, ticketMessageService),
	}
}

// Legacy compatibility layer - maintains the same interface as the original AdminTicketHandler
// This allows existing code to continue working without changes while using the modular structure internally

// ListTickets delegates to the list module
func (m *AdminTicketManager) ListTickets(c *gin.Context) {
	m.List.ListTickets(c)
}

// GetTicket delegates to the detail module
func (m *AdminTicketManager) GetTicket(c *gin.Context) {
	m.Detail.GetTicket(c)
}

// GetTicketByNumber delegates to the detail module
func (m *AdminTicketManager) GetTicketByNumber(c *gin.Context) {
	m.Detail.GetTicketByNumber(c)
}

// UpdateTicket delegates to the management module
func (m *AdminTicketManager) UpdateTicket(c *gin.Context) {
	m.Management.UpdateTicket(c)
}

// AssignTicket delegates to the status module
func (m *AdminTicketManager) AssignTicket(c *gin.Context) {
	m.Status.AssignTicket(c)
}

// ResolveTicket delegates to the status module
func (m *AdminTicketManager) ResolveTicket(c *gin.Context) {
	m.Status.ResolveTicket(c)
}

// CloseTicket delegates to the status module
func (m *AdminTicketManager) CloseTicket(c *gin.Context) {
	m.Status.CloseTicket(c)
}

// DeleteTicket delegates to the actions module
func (m *AdminTicketManager) DeleteTicket(c *gin.Context) {
	m.Actions.DeleteTicket(c)
}

// AddTicketMessage delegates to the messages module
func (m *AdminTicketManager) AddTicketMessage(c *gin.Context) {
	m.Messages.AddTicketMessage(c)
}

// GetTicketMessages delegates to the messages module
func (m *AdminTicketManager) GetTicketMessages(c *gin.Context) {
	m.Messages.GetTicketMessages(c)
}

// GetTicketStats delegates to the stats module
func (m *AdminTicketManager) GetTicketStats(c *gin.Context) {
	m.Stats.GetTicketStats(c)
}

// SetupAdminTicketRoutes sets up routes for admin ticket operations
func (m *AdminTicketManager) SetupAdminTicketRoutes(adminGroup *gin.RouterGroup) {
	// Note: Authentication and admin middleware are already applied by parent group
	// adminGroup.Use(middleware.AuthMiddleware(authService))
	// adminGroup.Use(middleware.RequireAdmin())

	// Ticket routes
	ticketGroup := adminGroup.Group("/tickets")
	{
		// Basic CRUD operations
		ticketGroup.GET("", m.ListTickets)
		ticketGroup.GET("/:id", m.GetTicket)
		ticketGroup.PUT("/:id", m.UpdateTicket)
		ticketGroup.DELETE("/:id", m.DeleteTicket)
		ticketGroup.GET("/number/:ticket_no", m.GetTicketByNumber)
		
		// Ticket management operations
		ticketGroup.POST("/:id/assign", m.AssignTicket)
		ticketGroup.POST("/:id/resolve", m.ResolveTicket)
		ticketGroup.POST("/:id/close", m.CloseTicket)
		
		// Message operations
		ticketGroup.POST("/:id/messages", m.AddTicketMessage)
		ticketGroup.GET("/:id/messages", m.GetTicketMessages)
		
		// Statistics
		ticketGroup.GET("/stats", m.GetTicketStats)
	}
}