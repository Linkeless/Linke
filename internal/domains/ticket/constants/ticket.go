package constants

// Ticket status constants
const (
	TicketStatusOpen       = "open"
	TicketStatusInProgress = "in_progress"
	TicketStatusPending    = "pending"
	TicketStatusResolved   = "resolved"
	TicketStatusClosed     = "closed"
)

// Ticket priority constants
const (
	TicketPriorityLow      = "low"
	TicketPriorityNormal   = "normal"
	TicketPriorityHigh     = "high"
	TicketPriorityUrgent   = "urgent"
	TicketPriorityCritical = "critical"
)

// Ticket category constants
const (
	TicketCategoryGeneral      = "general"
	TicketCategoryTechnical    = "technical"
	TicketCategoryBilling      = "billing"
	TicketCategoryAccount      = "account"
	TicketCategoryFeature      = "feature"
	TicketCategoryBug          = "bug"
	TicketCategorySubscription = "subscription"
	TicketCategoryPayment      = "payment"
)

// Ticket message type constants
const (
	MessageTypeUser   = "user"
	MessageTypeAdmin  = "admin"
	MessageTypeSystem = "system"
)

// Table name constants
const (
	TableTickets        = "tickets"
	TableTicketMessages = "ticket_messages"
)

// Ticket number format constants
const (
	TicketNumberPrefix = "TK"
	TicketNumberLength = 6
)

// Business logic constants
const (
	DefaultAgentMaxLoad        = 10
	DefaultAgentResponseTime   = 30  // minutes
	DefaultSatisfactionScore   = 8.5
	AutoAssignmentEnabled      = true
)

// Validation constraints
const (
	TicketTitleMinLength       = 5
	TicketTitleMaxLength       = 255
	TicketDescriptionMinLength = 10
	TicketDescriptionMaxLength = 5000
	MessageContentMinLength    = 1
	MessageContentMaxLength    = 5000
	ResolutionMinLength        = 10
	ResolutionMaxLength        = 5000
	EscalationReasonMinLength  = 10
	EscalationReasonMaxLength  = 1000
)

// Validation patterns
const (
	StatusValidationPattern   = "open in_progress pending resolved closed"
	PriorityValidationPattern = "low normal high urgent critical"
	CategoryValidationPattern = "general technical billing account feature bug subscription payment"
	MessageTypeValidationPattern = "user admin system"
)