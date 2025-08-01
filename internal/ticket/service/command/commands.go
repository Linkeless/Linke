package command

import (
	"linke/internal/ticket/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// CreateTicketCommand represents a command to create a ticket
type CreateTicketCommand struct {
	UserID      sharedvo.UserID
	Title       string
	Description string
	Category    string
	Priority    string
	Tags        []string
	Metadata    map[string]interface{}
}

// UpdateTicketCommand represents a command to update a ticket
type UpdateTicketCommand struct {
	TicketID    valueobject.TicketID
	UserID      sharedvo.UserID // User making the request
	IsAdmin     bool
	Title       *string
	Description *string
	Category    *string
	Priority    *string
	Tags        []string
}

// AssignTicketCommand represents a command to assign a ticket
type AssignTicketCommand struct {
	TicketID     valueobject.TicketID
	AssignedToID sharedvo.UserID
	AssignedBy   sharedvo.UserID // User making the assignment
}

// ChangeTicketStatusCommand represents a command to change ticket status
type ChangeTicketStatusCommand struct {
	TicketID  valueobject.TicketID
	NewStatus string
	ChangedBy sharedvo.UserID
	IsAdmin   bool
}

// ResolveTicketCommand represents a command to resolve a ticket
type ResolveTicketCommand struct {
	TicketID   valueobject.TicketID
	ResolvedBy sharedvo.UserID
	Resolution string
}

// CloseTicketCommand represents a command to close a ticket
type CloseTicketCommand struct {
	TicketID valueobject.TicketID
	ClosedBy sharedvo.UserID
	IsAdmin  bool
}

// AddTicketMessageCommand represents a command to add a message to a ticket
type AddTicketMessageCommand struct {
	TicketID    valueobject.TicketID
	UserID      sharedvo.UserID
	Content     string
	MessageType string
	Attachments []AttachmentData
	IsInternal  bool
	Metadata    map[string]interface{}
}

// AttachmentData represents attachment data in commands
type AttachmentData struct {
	Name string
	URL  string
	Size int64
	Type string
}