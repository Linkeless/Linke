package telegram

// TicketNotificationType represents the type of ticket notification
type TicketNotificationType string

const (
	TicketNotificationTypeCreated       TicketNotificationType = "created"
	TicketNotificationTypeAssigned      TicketNotificationType = "assigned"
	TicketNotificationTypeStatusChanged TicketNotificationType = "status_changed"
	TicketNotificationTypeReplied       TicketNotificationType = "replied"
	TicketNotificationTypeResolved      TicketNotificationType = "resolved"
	TicketNotificationTypeClosed        TicketNotificationType = "closed"
	TicketNotificationTypeEscalated     TicketNotificationType = "escalated"
)

// TicketNotification contains all the information needed for a ticket notification
type TicketNotification struct {
	Type        TicketNotificationType `json:"type"`
	TicketID    uint                   `json:"ticket_id"`
	TicketNo    string                 `json:"ticket_no"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Priority    string                 `json:"priority"`
	Status      string                 `json:"status"`
	OldStatus   string                 `json:"old_status,omitempty"`

	// User information
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`

	// Assignment information
	AssignedToID   uint   `json:"assigned_to_id,omitempty"`
	AssignedToName string `json:"assigned_to_name,omitempty"`

	// Reply information
	RepliedByID   uint   `json:"replied_by_id,omitempty"`
	RepliedByName string `json:"replied_by_name,omitempty"`
	ReplyContent  string `json:"reply_content,omitempty"`

	// Resolution information
	Resolution   string `json:"resolution,omitempty"`
	ClosedReason string `json:"closed_reason,omitempty"`

	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewTicketNotification creates a new ticket notification
func NewTicketNotification(notificationType TicketNotificationType) *TicketNotification {
	return &TicketNotification{
		Type:     notificationType,
		Metadata: make(map[string]interface{}),
	}
}
